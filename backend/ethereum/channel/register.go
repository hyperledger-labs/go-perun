// Copyright 2020 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package channel

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/channel"
)

// Register registers a state on-chain.
// If the state is a final state, register becomes a no-op.
func (a *Adjudicator) Register(ctx context.Context, req channel.AdjudicatorReq) (*channel.RegisteredEvent, error) {
	if req.Tx.State.IsFinal {
		return a.registerFinal(ctx, req)
	}
	return a.registerNonFinal(ctx, req)
}

// registerFinal registers a final state. It ensures that the final state is
// concluded on the adjudicator conctract.
func (a *Adjudicator) registerFinal(ctx context.Context, req channel.AdjudicatorReq) (*channel.RegisteredEvent, error) {
	// In the case of final states, we already call concludeFinal on the
	// adjudicator. Method ensureConcluded calls concludeFinal for final states.
	if err := a.ensureConcluded(ctx, req); err != nil {
		return nil, errors.WithMessage(err, "ensuring Concluded")
	}

	return channel.NewRegisteredEvent(
		req.Params.ID(),
		new(channel.ElapsedTimeout), // concludeFinal skips registration
		req.Tx.Version,
	), nil
}

func (a *Adjudicator) registerNonFinal(ctx context.Context, req channel.AdjudicatorReq) (*channel.RegisteredEvent, error) {
	_sub, err := a.Subscribe(ctx, req.Params)
	if err != nil {
		return nil, err
	}
	sub := _sub.(*RegisteredSub)
	// nolint:errcheck
	defer sub.Close()

	// call register if there was no past event
	if !sub.hasPast() {
		if err := a.callRegister(ctx, req); IsErrTxFailed(err) {
			a.log.Warn("Calling register failed, waiting for event anyways...")
		} else if err != nil {
			return nil, errors.WithMessage(err, "calling register")
		}
	}

	// iterate over state registrations and call refute until correct version got
	// registered.
	for {
		switch ev := sub.Next().(type) {
		case nil:
			// the subscription error might be nil, so to ensure a non-nil error, we
			// create a new one.
			return nil, errors.Errorf("subscription closed with error %v", sub.Err())

		case *channel.RegisteredEvent:
			if req.Tx.Version > ev.Version {
				if err := a.callRefute(ctx, req); IsErrTxFailed(err) {
					a.log.Warn("Calling refute failed, waiting for event anyways...")
				} else if err != nil {
					return nil, errors.WithMessage(err, "calling refute")
				}
				continue // wait for next event
			}
			return ev, nil // version matches, we're done

		case *channel.ProgressedEvent:
			return nil, errors.New("refutation phase already finished")
		}
	}
}

// Subscribe returns a new AdjudicatorSubscription to adjudicator events.
func (a *Adjudicator) Subscribe(ctx context.Context, params *channel.Params) (channel.AdjudicatorSubscription, error) {
	stored := make(chan *adjudicator.AdjudicatorStored)
	sub, iter, err := a.filterWatchStored(ctx, stored, params)
	if err != nil {
		return nil, errors.WithMessage(err, "filter/watch Stored event")
	}

	rsub := &RegisteredSub{
		cr:   a.ContractInterface,
		sub:  sub,
		next: make(chan *channel.RegisteredEvent, 1),
		err:  make(chan error, 1),
	}

	// Start event updater routine
	go rsub.updateNext(stored)

	// find past event, if any
	var ev *adjudicator.AdjudicatorStored
	for iter.Next() {
		ev = iter.Event // fast-forward to newest event
	}
	// nolint:errcheck,gosec,gosec
	iter.Close()
	if err := iter.Error(); err != nil {
		sub.Unsubscribe()
		return nil, errors.Wrap(err, "event iterator")
	}
	// Pass non-nil past event to updater
	if ev != nil {
		rsub.past = true
		stored <- ev
	}

	return rsub, nil
}

// filterWatchStored sets up a filter and a subscription on Stored events.
func (a *Adjudicator) filterWatchStored(ctx context.Context, stored chan *adjudicator.AdjudicatorStored, params *channel.Params) (sub event.Subscription, iter *adjudicator.AdjudicatorStoredIterator, err error) {
	defer func() {
		if err != nil && sub != nil {
			sub.Unsubscribe()
		}
	}()
	// Watch new events
	watchOpts, err := a.NewWatchOpts(ctx)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "creating watchopts")
	}
	sub, err = a.contract.WatchStored(watchOpts, stored, []channel.ID{params.ID()})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "watching stored events")
	}

	// Filter old Events
	filterOpts, err := a.NewFilterOpts(ctx)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "creating filter opts")
	}
	iter, err = a.contract.FilterStored(filterOpts, []channel.ID{params.ID()})
	if err != nil {
		return nil, nil, errors.Wrap(err, "filtering stored events")
	}

	return sub, iter, nil
}

// RegisteredSub implements the channel.RegisteredSubscription interface.
type RegisteredSub struct {
	cr   ethereum.ChainReader          // chain reader to read block time
	sub  event.Subscription            // Stored event subscription
	next chan *channel.RegisteredEvent // Registered event sink
	err  chan error                    // error from subscription
	past bool                          // whether there was a past event when the subscription was created
}

func (r *RegisteredSub) hasPast() bool {
	return r.past
}

func (r *RegisteredSub) updateNext(events chan *adjudicator.AdjudicatorStored) {
evloop:
	for {
		select {
		case next := <-events:
			select {
			// drain next-channel on new event
			case current := <-r.next:
				currentTimeout := current.Timeout().(*BlockTimeout)
				// if newer version or same version and newer timeout, replace
				if current.Version < next.Version ||
					current.Version == next.Version && currentTimeout.Time < next.Timeout {
					r.next <- r.storedToRegisteredEvent(next)
				} else { // otherwise, reuse old
					r.next <- current
				}
			default: // next-channel is empty
				r.next <- r.storedToRegisteredEvent(next)
			}
		case err := <-r.sub.Err():
			r.err <- err
			break evloop
		}
	}

	// subscription got closed, close next channel and return
	select {
	case <-r.next:
	default:
	}
	close(r.next)
}

// Next returns the newest past or next blockchain event.
// It blocks until an event is returned from the blockchain or the subscription
// is closed. If the subscription is closed, Next immediately returns nil.
// If there was a past event when the subscription was set up, the first call to
// Next will return it.
func (r *RegisteredSub) Next() channel.AdjudicatorEvent {
	reg := <-r.next
	if reg == nil {
		return nil // otherwise we get (*RegisteredEvent)(nil)
	}
	return reg
}

// Close closes this subscription. Any pending calls to Next will return nil.
func (r *RegisteredSub) Close() error {
	r.sub.Unsubscribe()
	return nil
}

// Err returns the error of the event subscription.
// Should only be called after Next returned nil.
func (r *RegisteredSub) Err() error {
	return <-r.err
}

func (r *RegisteredSub) storedToRegisteredEvent(event *adjudicator.AdjudicatorStored) *channel.RegisteredEvent {
	if event == nil {
		return nil
	}
	return channel.NewRegisteredEvent(
		event.ChannelID,
		NewBlockTimeout(r.cr, event.Timeout),
		event.Version,
	)
}
