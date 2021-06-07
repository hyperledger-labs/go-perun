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
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	cherrors "perun.network/go-perun/backend/ethereum/channel/errors"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
)

// Subscribe returns a new AdjudicatorSubscription to adjudicator events.
func (a *Adjudicator) Subscribe(ctx context.Context, params *channel.Params) (channel.AdjudicatorSubscription, error) {
	events := make(chan *adjudicator.AdjudicatorChannelUpdate)
	sub, iter, err := a.filterWatch(ctx, events, params)
	if err != nil {
		return nil, errors.WithMessage(err, "creating filter-watch event subscription")
	}

	rsub := &RegisteredSub{
		cr:   a.ContractInterface,
		sub:  sub,
		next: make(chan channel.AdjudicatorEvent, 1),
		err:  make(chan error, 1),
	}

	// Start event updater routine
	go rsub.updateNext(ctx, events, a)

	// find past event, if any
	var ev *adjudicator.AdjudicatorChannelUpdate
	for iter.Next() {
		ev = iter.Event // fast-forward to newest event
	}
	// nolint:errcheck,gosec,gosec
	iter.Close()
	if err := iter.Error(); err != nil {
		err = cherrors.CheckIsChainNotReachableError(err)
		sub.Unsubscribe()
		return nil, errors.WithMessage(err, "event iterator")
	}
	// Pass non-nil past event to updater
	if ev != nil {
		events <- ev
	}

	return rsub, nil
}

// filterWatch sets up a filter and a subscription on events.
func (a *Adjudicator) filterWatch(ctx context.Context, events chan *adjudicator.AdjudicatorChannelUpdate, params *channel.Params) (sub event.Subscription, iter *adjudicator.AdjudicatorChannelUpdateIterator, err error) {
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
	sub, err = a.contract.WatchChannelUpdate(watchOpts, events, []channel.ID{params.ID()})
	if err != nil {
		err = cherrors.CheckIsChainNotReachableError(err)
		return nil, nil, errors.WithMessagef(err, "watching events")
	}

	// Filter old Events
	filterOpts, err := a.NewFilterOpts(ctx)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "creating filter opts")
	}
	iter, err = a.contract.FilterChannelUpdate(filterOpts, []channel.ID{params.ID()})
	if err != nil {
		err = cherrors.CheckIsChainNotReachableError(err)
		return nil, nil, errors.WithMessage(err, "filtering events")
	}

	return sub, iter, nil
}

// RegisteredSub implements the channel.RegisteredSubscription interface.
type RegisteredSub struct {
	cr     ethereum.ChainReader          // chain reader to read block time
	sub    event.Subscription            // Event subscription
	next   chan channel.AdjudicatorEvent // Event sink
	err    chan error                    // error from subscription
	closed sync.Once
}

func (r *RegisteredSub) updateNext(ctx context.Context, events chan *adjudicator.AdjudicatorChannelUpdate, a *Adjudicator) {
evloop:
	for {
		select {
		case next := <-events:
			select {
			// drain next-channel on new event
			case current := <-r.next:
				currentTimeout := current.Timeout().(*BlockTimeout)
				// if newer version or same version and newer timeout, replace
				if current.Version() < next.Version || current.Version() == next.Version && currentTimeout.Time < next.Timeout {
					e, err := a.convertEvent(ctx, next)
					if err != nil {
						r.err <- err
						break evloop
					}

					r.next <- e
				} else { // otherwise, reuse old
					r.next <- current
				}
			default: // next-channel is empty
				e, err := a.convertEvent(ctx, next)
				if err != nil {
					r.err <- err
					break evloop
				}

				r.next <- e
			}
		case err := <-r.sub.Err():
			r.err <- cherrors.CheckIsChainNotReachableError(err)
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
	r.closed.Do(r.sub.Unsubscribe)
	return nil
}

// Err returns the error of the event subscription.
// Should only be called after Next returned nil.
func (r *RegisteredSub) Err() error {
	return <-r.err
}

func (a *Adjudicator) convertEvent(ctx context.Context, e *adjudicator.AdjudicatorChannelUpdate) (channel.AdjudicatorEvent, error) {
	base := channel.NewAdjudicatorEventBase(e.ChannelID, NewBlockTimeout(a.ContractInterface, e.Timeout), e.Version)
	switch e.Phase {
	case phaseDispute:
		return &channel.RegisteredEvent{AdjudicatorEventBase: *base}, nil

	case phaseForceExec:
		args, err := a.fetchProgressCallData(ctx, e.Raw.TxHash)
		if err != nil {
			return nil, errors.WithMessage(err, "fetching call data")
		}
		app, err := channel.Resolve(wallet.AsWalletAddr(args.Params.App))
		if err != nil {
			return nil, errors.WithMessage(err, "resolving app")
		}
		newState := FromEthState(app, &args.State)
		return &channel.ProgressedEvent{
			AdjudicatorEventBase: *base,
			State:                &newState,
			Idx:                  channel.Index(args.ActorIdx.Uint64()),
		}, nil

	case phaseConcluded:
		return &channel.ConcludedEvent{AdjudicatorEventBase: *base}, nil

	default:
		panic("unknown phase")
	}
}

type progressCallData struct {
	Params   adjudicator.ChannelParams
	StateOld adjudicator.ChannelState
	State    adjudicator.ChannelState
	ActorIdx *big.Int
	Sig      []byte
}

func (a *Adjudicator) fetchProgressCallData(ctx context.Context, txHash common.Hash) (*progressCallData, error) {
	tx, _, err := a.ContractBackend.TransactionByHash(ctx, txHash)
	if err != nil {
		err = cherrors.CheckIsChainNotReachableError(err)
		return nil, errors.WithMessage(err, "getting transaction")
	}

	argsData := tx.Data()[len(abiProgress.ID):]

	argsI, err := abiProgress.Inputs.UnpackValues(argsData)
	if err != nil {
		return nil, errors.WithMessage(err, "unpacking")
	}

	var args progressCallData
	err = abiProgress.Inputs.Copy(&args, argsI)
	if err != nil {
		return nil, errors.WithMessage(err, "copying into struct")
	}

	return &args, nil
}
