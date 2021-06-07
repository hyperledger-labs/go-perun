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

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	cherrors "perun.network/go-perun/backend/ethereum/channel/errors"
)

type (
	// EventSub generic event subscription.
	// Can be used on any Contract with any Event.
	// This EventSub does not prevent duplicates.
	EventSub struct {
		closed   chan struct{}
		contract *bind.BoundContract
		eFact    EventFactory

		watchLogs, filterLogs chan types.Log
		watchSub, filterSub   event.Subscription
	}

	// Event is a generic on-chain event.
	Event struct {
		Name   string      // Name of the event. Must match the ABI definition.
		Data   interface{} // Instance of the concrete Event type.
		Filter Filter      // Filters Events by their body.
		Log    types.Log   // Raw original log for additional information.
	}

	// EventFactory is used to create `Event`s.
	// The `Data` and `Name` fields must be set. `Filter` is optional.
	EventFactory func() *Event

	// Filter can be used to filter events.
	// Look at `TestEventSub_Filter` test or the auto generated Filter- and
	// Watch-functions in the bindings/ folder for an example.
	Filter [][]interface{}
)

// NewEventSub creates a new `EventSub`. Should always be closed with `Close`.
// `pastBlocks` can be used to define how many blocks into the past the sub
// should query.
func NewEventSub(ctx context.Context, chain ethereum.ChainReader, contract *bind.BoundContract, eFact EventFactory, pastBlocks uint64) (*EventSub, error) {
	// Get start block number.
	startBlock, err := calcStartBlock(ctx, chain, pastBlocks)
	if err != nil {
		return nil, errors.WithMessage(err, "calculating starting block number")
	}
	// Watch for future events.
	event := eFact()
	watchOpts := &bind.WatchOpts{Start: &startBlock}
	watchLogs, watchSub, err := contract.WatchLogs(watchOpts, event.Name, event.Filter...)
	if err != nil {
		err = cherrors.CheckIsChainNotReachableError(err)
		return nil, errors.WithMessage(err, "watching logs")
	}
	// Read past events.
	filterOpts := &bind.FilterOpts{Start: startBlock}
	filterLogs, filterSub, err := contract.FilterLogs(filterOpts, event.Name, event.Filter...)
	if err != nil {
		watchSub.Unsubscribe()
		err = cherrors.CheckIsChainNotReachableError(err)
		return nil, errors.WithMessage(err, "filtering logs")
	}

	return &EventSub{
		closed:     make(chan struct{}),
		contract:   contract,
		eFact:      eFact,
		watchLogs:  watchLogs,
		filterLogs: filterLogs,
		watchSub:   watchSub,
		filterSub:  filterSub,
	}, nil
}

func calcStartBlock(ctx context.Context, chain ethereum.ChainReader, pastBlocks uint64) (uint64, error) {
	current, err := chain.HeaderByNumber(ctx, nil)
	if err != nil {
		err = cherrors.CheckIsChainNotReachableError(err)
		return 0, errors.WithMessage(err, "retrieving latest block")
	}
	if current.Number.Uint64() <= pastBlocks {
		return 1, nil
	}
	return current.Number.Uint64() - pastBlocks, nil
}

// Read reads all past and future events into `sink`.
// Can be aborted by cancelling `ctx` or `Close()`.
// It is possible that the same event is read more than once.
// After casting the generic event to a specific type, the `Raw` log field
// will be nil. Use the `Log` field of the generic event instead.
func (s *EventSub) Read(ctx context.Context, sink chan<- *Event) error {
	// First read into the past.
	if err := s.readPast(ctx, sink); err != nil {
		return errors.WithMessage(err, "reading logs")
	}
	// Then wait for new events.
	if err := s.readFuture(ctx, sink); err != nil {
		return errors.WithMessage(err, "reading logs")
	}
	return nil
}

// ReadPast reads all past events into `sink`.
// Can be aborted by cancelling `ctx` or `Close()`.
// It is possible that the same event is read more than once.
// After casting the generic event to a specific type, the `Raw` log field
// will be nil. Use the `Log` field of the generic event instead.
func (s *EventSub) ReadPast(ctx context.Context, sink chan<- *Event) error {
	return errors.WithMessage(s.readPast(ctx, sink), "reading logs")
}

func (s *EventSub) readPast(ctx context.Context, sink chan<- *Event) error {
	var logs []types.Log
	// Two read loops are needed if the event sub is closed before all events
	// could be read.
read1:
	for {
		select {
		case log := <-s.filterLogs:
			logs = append(logs, log)
		case err := <-s.filterSub.Err():
			if err != nil {
				err = cherrors.CheckIsChainNotReachableError(err)
				return err
			}
			break read1
		case <-ctx.Done():
			return ctx.Err()
		case <-s.closed:
			return nil
		}
	}
read2:
	for {
		select {
		case log := <-s.filterLogs:
			logs = append(logs, log)
		case <-s.closed:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			break read2
		}
	}

	for _, log := range logs {
		event := s.eFact()
		if err := s.contract.UnpackLog(event.Data, event.Name, log); err != nil {
			return err
		}
		event.Log = log

		select {
		case <-ctx.Done():
			return ctx.Err()
		case sink <- event:
		case <-s.closed:
			return nil
		}
	}
	return nil
}

func (s *EventSub) readFuture(ctx context.Context, sink chan<- *Event) error {
	for {
		select {
		case log := <-s.watchLogs:
			event := s.eFact()
			if err := s.contract.UnpackLog(event.Data, event.Name, log); err != nil {
				return err
			}
			event.Log = log

			select {
			case <-ctx.Done():
				return ctx.Err()
			case sink <- event:
			case <-s.closed:
				return nil
			}
		case err := <-s.watchSub.Err():
			err = cherrors.CheckIsChainNotReachableError(err)
			return err
		case <-ctx.Done():
			return ctx.Err()
		case <-s.closed:
			return nil
		}
	}
}

// Close closes the sub and frees associated resources.
// Should be called exactly once and panics otherwise.
// Must not be called if the construction function returned with an error.
func (s *EventSub) Close() {
	close(s.closed)
	s.watchSub.Unsubscribe()
	s.filterSub.Unsubscribe()
}
