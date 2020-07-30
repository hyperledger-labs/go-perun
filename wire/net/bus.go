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

package net

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

// Bus implements the wire.Bus interface using network connections.
type Bus struct {
	reg      *EndpointRegistry
	mainRecv *wire.Receiver
	recvs    map[wallet.AddrKey]wire.Consumer
	mutex    sync.RWMutex // Protects reg, recv.
}

// NewBus creates a new network bus. The dialer and listener are used to
// establish new connections internally, while id is this node's identity.
func NewBus(id wire.Account, d Dialer) *Bus {
	b := &Bus{
		mainRecv: wire.NewReceiver(),
		recvs:    make(map[wallet.AddrKey]wire.Consumer),
	}

	onNewEndpoint := func(wire.Address) wire.Consumer { return b.mainRecv }
	b.reg = NewEndpointRegistry(id, onNewEndpoint, d)
	go b.dispatchMsgs()

	return b
}

// Listen listens for incoming connections to add to the Bus.
func (b *Bus) Listen(l Listener) {
	b.reg.Listen(l)
}

// SubscribeClient subscribes a new client to the bus. Duplicate subscriptions
// are forbidden and will cause a panic. The supplied consumer will receive all
// messages that are sent to the requested address.
func (b *Bus) SubscribeClient(c wire.Consumer, addr wire.Address) error {
	b.addSubscriber(c, addr)
	c.OnCloseAlways(func() { b.removeSubscriber(addr) })
	return nil
}

// Publish sends an envelope to its recipient. Automatically establishes a
// communication channel to the recipient using the bus' dialer. Only returns
// when the context is aborted or the envelope was sent successfully.
func (b *Bus) Publish(ctx context.Context, e *wire.Envelope) (err error) {
	for {
		var ep *Endpoint
		if ep, err = b.reg.Get(ctx, e.Recipient); err == nil {
			if err = ep.Send(ctx, e); err == nil {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return errors.WithMessagef(err, "publishing %T envelope", e.Msg)
		case <-b.ctx().Done():
			return errors.Errorf("publishing %T envelope: Bus closed", e.Msg)
		default:
		}
	}
}

// Close closes the bus and terminates its goroutines.
func (b *Bus) Close() error {
	if err := b.mainRecv.Close(); err != nil {
		return err
	}

	b.mutex.Lock()
	b.recvs = nil
	b.mutex.Unlock()

	return b.reg.Close()
}

func (b *Bus) addSubscriber(c wire.Consumer, addr wire.Address) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, ok := b.recvs[wallet.Key(addr)]; ok {
		log.Panic("duplicate SubscribeClient")
	}

	b.recvs[wallet.Key(addr)] = c
}

// ctx returns the context of the bus' registry.
func (b *Bus) ctx() context.Context {
	return b.reg.Ctx()
}

// dispatchMsgs dispatches all received messages to their subscribed clients.
func (b *Bus) dispatchMsgs() {
	for {
		// Return when the bus is closed.
		e, err := b.mainRecv.Next(b.ctx())
		if err != nil {
			return
		}

		b.mutex.Lock()
		r, ok := b.recvs[wallet.Key(e.Recipient)]
		b.mutex.Unlock()
		if !ok {
			log.WithField("sender", e.Sender).
				WithField("recipient", e.Recipient).
				Warnf("Received %T message for unknown recipient", e.Msg)
		} else {
			r.Put(e)
		}
	}
}

func (b *Bus) removeSubscriber(addr wire.Address) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, ok := b.recvs[wallet.Key(addr)]; !ok {
		log.Panic("deleting nonexisting subscriber")
	}

	delete(b.recvs, wallet.Key(addr))
}
