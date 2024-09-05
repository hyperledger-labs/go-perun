// Copyright 2022 - See NOTICE file for copyright holders.
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

package wire

import (
	"context"
	"fmt"
	"perun.network/go-perun/wallet"

	"polycry.pt/poly-go/errors"
	"polycry.pt/poly-go/sync/atomic"
)

var _ Bus = (*hybridBus)(nil)

// A Bus that can send messages over a local bus and a remote bus.
type hybridBus struct {
	buses []Bus
}

// NewHybridBus creates a hybrid bus that sends and receives messages over
// multiple buses. Publishing a message over a hybrid bus will attempt to
// publish it over all its sub-buses simultaneously. Subscribing a client to the
// hybrid bus will subscribe it to all its sub-buses. The sub-buses can still be
// accessed independently for more fine-grained control.
//
// Once any bus indicates success in publishing a message, the publishing
// operations on all other buses are cancelled. This bus is best used with a
// LocalBus and one network bus, otherwise, with multiple non-instantaneous
// buses, messages might get sent multiple times. Buses that indicate success
// immediately (such as when caching outbound messages) can also cause problems
// if they cancel non-immediate buses.
func NewHybridBus(buses ...Bus) Bus {
	if len(buses) == 0 {
		panic("expected at least one sub-bus.")
	}
	for i, bus := range buses {
		if bus == nil {
			panic(fmt.Sprintf("nil Bus[%d] passed to NewHybridBus", i))
		}
	}

	return &hybridBus{buses: buses}
}

// Publish an envelope over all sub-buses simultaneously.
func (b *hybridBus) Publish(ctx context.Context, e *Envelope) error {
	// Whether an attempt to publish the envelope succeeded.
	var success atomic.Bool
	// Sub-context for sending, cancel on first success.
	sending, cancelSending := context.WithCancel(ctx)
	defer cancelSending()

	errg := errors.NewGatherer()
	for _, bus := range b.buses {
		bus := bus
		errg.Go(func() error {
			err := bus.Publish(sending, e)
			// If sending was successful, abort all other send operations, and
			// indicate success.
			if err == nil {
				success.Set()
				cancelSending()
			}
			return err
		})
	}

	// Wait until all sending attempts terminated. This is when either the
	// context expires, all buses fail, or at least one bus succeeds.
	err := errg.Wait()
	if success.IsSet() {
		return nil
	}

	return err
}

// SubscribeClient subscribes an envelope consumer to all sub-buses.
func (b *hybridBus) SubscribeClient(c Consumer, receiver map[wallet.BackendID]Address) error {
	errg := errors.NewGatherer()
	for _, bus := range b.buses {
		errg.Add(bus.SubscribeClient(c, receiver))
	}
	return errg.Err()
}
