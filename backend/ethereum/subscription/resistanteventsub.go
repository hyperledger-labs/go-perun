// Copyright 2021 - See NOTICE file for copyright holders.
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

package subscription

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	cherrors "perun.network/go-perun/backend/ethereum/channel/errors"
	"perun.network/go-perun/log"
	pkgsync "polycry.pt/poly-go/sync"
)

type (
	// ResistantEventSub wraps an `EventSub` and makes it resistant to chain reorgs.
	// It handles `removed` and `rebirth` events and has a `finalityDepth`
	// threshold to decide when an Event is final.
	// It will never emit the same event twice.
	ResistantEventSub struct {
		closer        pkgsync.Closer
		sub           *EventSub
		finalityDepth *big.Int

		lastBlockNum *big.Int
		heads        chan *types.Header
		headSub      ethereum.Subscription
		events       map[common.Hash]*Event
	}
)

const (
	// number of headers that a resistant event sub can buffer.
	resistantSubHeadBuffSize = 128
	// number of raw events that a resistant event sub can buffer.
	resistantSubRawEventBuffSize = 128
)

// Subscribe is a convenience function which returns a `ResistantEventSub`.
// It is equivalent to manually calling `NewEventSub` and `NewResistantEventSub`
// with the given parameters.
func Subscribe(ctx context.Context, cr ethereum.ChainReader, contract *bind.BoundContract, eFact EventFactory, startBlockOffset, confirmations uint64) (*ResistantEventSub, error) {
	_sub, err := NewEventSub(ctx, cr, contract, eFact, startBlockOffset)
	if err != nil {
		return nil, errors.WithMessage(err, "creating filter-watch event subscription")
	}
	sub, err := NewResistantEventSub(ctx, _sub, cr, confirmations)
	if err != nil {
		_sub.Close()
		return nil, errors.WithMessage(err, "creating filter-watch event subscription")
	}
	return sub, nil
}

// NewResistantEventSub creates a new `ResistantEventSub` from the given
// `EventSub`. Closes the passed `EventSub` when done.
// `finalityDepth` defines in how many blocks an event needs to be included.
// `finalityDepth` cannot be smaller than 1.
// The passed `EventSub` should query more than `finalityDepth` blocks into
// the past.
func NewResistantEventSub(ctx context.Context, sub *EventSub, cr ethereum.ChainReader, finalityDepth uint64) (*ResistantEventSub, error) {
	if finalityDepth < 1 {
		panic("finalityDepth needs to be at least 1")
	}
	last, err := cr.HeaderByNumber(ctx, nil)
	if err != nil {
		err = cherrors.CheckIsChainNotReachableError(err)
		return nil, errors.WithMessage(err, "subscribing to headers")
	}
	log.Debugf("Resistant Event sub started at block: %v", last.Number)
	// Use a large buffer to not block geth.
	heads := make(chan *types.Header, resistantSubHeadBuffSize)
	headSub, err := cr.SubscribeNewHead(ctx, heads)
	if err != nil {
		headSub.Unsubscribe()
		err = cherrors.CheckIsChainNotReachableError(err)
		return nil, errors.WithMessage(err, "subscribing to headers")
	}

	fd := new(big.Int)
	fd.SetUint64(finalityDepth)
	ret := &ResistantEventSub{
		sub:           sub,
		lastBlockNum:  new(big.Int).Set(last.Number),
		heads:         heads,
		headSub:       headSub,
		finalityDepth: fd,
		events:        make(map[common.Hash]*Event),
	}
	ret.closer.OnCloseAlways(func() {
		headSub.Unsubscribe()
		sub.Close()
	})
	return ret, nil
}

// Read reads all past and future events into `sink`.
// Can be aborted by cancelling `ctx` or `Close()`.
// All events can be considered final.
func (s *ResistantEventSub) Read(_ctx context.Context, sink chan<- *Event) error {
	ctx, cancel := context.WithCancel(_ctx)
	defer cancel()
	subErr := make(chan error, 1)
	rawEvents := make(chan *Event, resistantSubRawEventBuffSize)
	// Read events from the underlying event subscription.
	go func() {
		subErr <- s.sub.Read(ctx, rawEvents)
	}()

	for {
		select {
		case head := <-s.heads:
			if head == nil {
				return errors.New("head sub returned nil")
			}
			s.processHead(head, sink)
		case event := <-rawEvents:
			s.processEvent(event, sink)
		case e := <-s.headSub.Err():
			return errors.WithMessage(e, "underlying head subscription")
		case e := <-subErr:
			if e != nil {
				return errors.WithMessage(e, "underlying EventSub.Read")
			}
			return errors.New("underlying event sub terminated")
		case <-ctx.Done():
			return ctx.Err()
		case <-s.closer.Closed():
			return nil
		}
	}
}

// ReadPast reads all past events into `sink`.
// Can be aborted by cancelling `ctx` or `Close()`.
// All events can be considered final.
func (s *ResistantEventSub) ReadPast(_ctx context.Context, sink chan<- *Event) error {
	ctx, cancel := context.WithCancel(_ctx)
	defer cancel()
	subErr := make(chan error, 1)
	rawEvents := make(chan *Event, resistantSubRawEventBuffSize)
	// Read events from the underlying event subscription.
	go func() {
		defer close(rawEvents)
		subErr <- s.sub.ReadPast(ctx, rawEvents)
	}()

	for {
		select {
		case head := <-s.heads:
			if head == nil {
				return errors.New("head sub returned nil")
			}
			s.processHead(head, sink)
		case event, ok := <-rawEvents:
			if !ok {
				s.drainHeadSub(sink)
				return errors.WithMessage(<-subErr, "underlying EventSub.Read")
			}
			s.processEvent(event, sink)
		case e := <-s.headSub.Err():
			return errors.WithMessage(e, "underlying head subscription")
		case <-ctx.Done():
			return ctx.Err()
		case <-s.closer.Closed():
			return nil
		}
	}
}

// drainHeadSub ensures that all queued block headers are processed.
func (s *ResistantEventSub) drainHeadSub(sink chan<- *Event) {
	for {
		select {
		case head := <-s.heads:
			s.processHead(head, sink)
		default:
			return
		}
	}
}

// handles events that are received from the geth node and checks if they
// are final.
func (s *ResistantEventSub) processEvent(event *Event, sink chan<- *Event) {
	hash := event.Log.TxHash
	log := log.WithField("hash", hash.Hex())

	if event.Log.Removed { //nolint:nestif
		if _, found := s.events[hash]; !found {
			log.Error("Race detected between event and header sub")
		} else {
			log.Trace("Event preliminary excluded")
			delete(s.events, hash)
		}
	} else {
		if s.isFinal(event) {
			sink <- event
			delete(s.events, hash)
		} else {
			log.Trace("Event preliminary included")
			s.events[hash] = event
		}
	}
}

// handles headers that are received from the geth node and checks if events
// become final.
func (s *ResistantEventSub) processHead(head *types.Header, sink chan<- *Event) {
	log.Tracef("Received new block. From %v to %v", s.lastBlockNum, head.Number)
	s.lastBlockNum.Set(head.Number)

	for _, event := range s.events {
		if s.isFinal(event) {
			sink <- event
			delete(s.events, event.Log.TxHash)
		}
	}
}

func (s *ResistantEventSub) isFinal(event *Event) bool {
	log := log.WithField("hash", event.Log.TxHash.Hex())
	diff := new(big.Int).Sub(s.lastBlockNum, big.NewInt(int64(event.Log.BlockNumber)))

	if diff.Sign() < 0 {
		log.Tracef("Event sub was faster than head sub, ignored")
		return false
	}

	included := new(big.Int).Add(diff, big.NewInt(1))
	if included.Cmp(s.finalityDepth) >= 0 {
		log.Debugf("Event final after %d block(s)", included)
		return true
	}

	log.Tracef("Event included %d time(s)", included)
	return false
}

// Close closes the sub and the underlying `EventSub`.
// Can be called more than once. Is thread safe.
func (s *ResistantEventSub) Close() {
	if err := s.closer.Close(); err != nil && !pkgsync.IsAlreadyClosedError(err) {
		log.WithError(err).Error("could not close EventSub")
	}
	// NOTE: The underlying `EventSub` is closed in the `OnCloseAlways` hook.
}
