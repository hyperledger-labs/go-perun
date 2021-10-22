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
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings"
	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	cherrors "perun.network/go-perun/backend/ethereum/channel/errors"
	"perun.network/go-perun/backend/ethereum/subscription"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
)

// Subscribe returns a new AdjudicatorSubscription to adjudicator events.
func (a *Adjudicator) Subscribe(ctx context.Context, chID channel.ID) (channel.AdjudicatorSubscription, error) {
	subErr := make(chan error, 1)
	events := make(chan *subscription.Event, 10)
	eFact := func() *subscription.Event {
		return &subscription.Event{
			Name:   bindings.Events.AdjChannelUpdate,
			Data:   new(adjudicator.AdjudicatorChannelUpdate),
			Filter: [][]interface{}{{chID}},
		}
	}
	sub, err := subscription.Subscribe(ctx, a.ContractBackend, a.bound, eFact, startBlockOffset, a.txFinalityDepth)
	if err != nil {
		return nil, errors.WithMessage(err, "creating filter-watch event subscription")
	}
	// Find new events
	go func() {
		subErr <- sub.Read(ctx, events)
	}()
	rsub := &RegisteredSub{
		cr:     a.ContractInterface,
		sub:    sub,
		subErr: subErr,
		next:   make(chan channel.AdjudicatorEvent, 1),
		err:    make(chan error, 1),
	}
	go rsub.updateNext(ctx, events, a)

	return rsub, nil
}

// RegisteredSub implements the channel.AdjudicatorSubscription interface.
type RegisteredSub struct {
	cr     ethereum.ChainReader            // chain reader to read block time
	sub    *subscription.ResistantEventSub // Event subscription
	subErr chan error
	next   chan channel.AdjudicatorEvent // Event sink
	err    chan error                    // error from subscription
}

func (r *RegisteredSub) updateNext(ctx context.Context, events chan *subscription.Event, a *Adjudicator) {
evloop:
	for {
		select {
		case _next := <-events:
			err := r.processNext(ctx, a, _next)
			if err != nil {
				r.err <- err
				break evloop
			}
		case err := <-r.subErr:
			if err != nil {
				r.err <- errors.WithMessage(err, "EventSub closed")
			} else {
				// Normal closing should produce no error
				close(r.err)
			}
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

func (r *RegisteredSub) processNext(ctx context.Context, a *Adjudicator, _next *subscription.Event) (err error) {
	next, ok := _next.Data.(*adjudicator.AdjudicatorChannelUpdate)
	next.Raw = _next.Log
	if !ok {
		log.Panicf("unexpected event type: %T", _next.Data)
	}

	select {
	// drain next-channel on new event
	case current := <-r.next:
		currentTimeout := current.Timeout().(*BlockTimeout)
		// if newer version or same version and newer timeout, replace
		if current.Version() < next.Version || current.Version() == next.Version && currentTimeout.Time < next.Timeout {
			var e channel.AdjudicatorEvent
			e, err = a.convertEvent(ctx, next)
			if err != nil {
				return
			}

			r.next <- e
		} else { // otherwise, reuse old
			r.next <- current
		}
	default: // next-channel is empty
		var e channel.AdjudicatorEvent
		e, err = a.convertEvent(ctx, next)
		if err != nil {
			return
		}

		r.next <- e
	}
	return
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
	r.sub.Close()
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
		args, err := a.fetchRegisterCallData(ctx, e.Raw.TxHash)
		if err != nil {
			return nil, errors.WithMessage(err, "fetching call data")
		}

		ch, ok := args.get(e.ChannelID)
		if !ok {
			return nil, errors.Errorf("channel not found in calldata: %v", e.ChannelID)
		}

		var app channel.App
		var zeroAddress common.Address
		if ch.Params.App == zeroAddress {
			app = channel.NoApp()
		} else {
			app, err = channel.Resolve(wallet.AsWalletAddr(ch.Params.App))
			if err != nil {
				return nil, err
			}
		}
		state := FromEthState(app, &ch.State)

		return &channel.RegisteredEvent{
			AdjudicatorEventBase: *base,
			State:                &state,
			Sigs:                 ch.Sigs,
		}, nil

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
	var args progressCallData
	err := a.fetchCallData(ctx, txHash, abiProgress, &args)
	return &args, errors.WithMessage(err, "fetching call data")
}

type registerCallData struct {
	Channel     adjudicator.AdjudicatorSignedState
	SubChannels []adjudicator.AdjudicatorSignedState
}

func (args *registerCallData) get(id channel.ID) (*adjudicator.AdjudicatorSignedState, bool) {
	ch := &args.Channel
	if ch.State.ChannelID == id {
		return ch, true
	}
	for _, ch := range args.SubChannels {
		if ch.State.ChannelID == id {
			return &ch, true
		}
	}
	return nil, false
}

func (a *Adjudicator) fetchRegisterCallData(ctx context.Context, txHash common.Hash) (*registerCallData, error) {
	var args registerCallData
	err := a.fetchCallData(ctx, txHash, abiRegister, &args)
	return &args, errors.WithMessage(err, "fetching call data")
}

func (a *Adjudicator) fetchCallData(ctx context.Context, txHash common.Hash, method abi.Method, args interface{}) error {
	tx, _, err := a.ContractBackend.TransactionByHash(ctx, txHash)
	if err != nil {
		err = cherrors.CheckIsChainNotReachableError(err)
		return errors.WithMessage(err, "getting transaction")
	}

	argsData := tx.Data()[len(method.ID):]

	argsI, err := method.Inputs.UnpackValues(argsData)
	if err != nil {
		return errors.WithMessage(err, "unpacking")
	}

	err = method.Inputs.Copy(args, argsI)
	if err != nil {
		return errors.WithMessage(err, "copying into struct")
	}

	return nil
}
