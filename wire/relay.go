// Copyright 2019 - See NOTICE file for copyright holders.
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
	stdsync "sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"polycry.pt/poly-go/sync"
)

// Relay handles (un)registering Consumers for a message Relay's messages.
type Relay struct {
	sync.Closer
	mutex     stdsync.RWMutex
	consumers []subscription

	cache             Cache
	defaultMsgHandler func(*Envelope) // Handles messages with no subscriber.
}

type subscription struct {
	consumer  Consumer
	predicate Predicate
}

// NewRelay returns a new Relay which logs unhandled messages.
func NewRelay() *Relay {
	return &Relay{defaultMsgHandler: logUnhandledMsg}
}

// Close closes the relay.
func (p *Relay) Close() error {
	if err := p.Closer.Close(); err != nil {
		return err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.consumers = nil

	cs := p.cache.Size()
	if cs != 0 {
		p.cache.Flush() // GC
		return errors.Errorf("cache was not empty (%d)", cs)
	}
	return nil
}

// Cache enables caching of messages that don't match any consumer. They are
// only cached if they match the given predicate, within the given context.
func (p *Relay) Cache(ctx context.Context, predicate Predicate) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.IsClosed() {
		return
	}

	p.cache.Cache(ctx, predicate)
}

// Subscribe adds a Consumer to the subscriptions.
// If the Consumer is already subscribed, Subscribe panics.
// If the producer is closed, Subscribe returns an error.
// Otherwise, Subscribe returns nil.
func (p *Relay) Subscribe(c Consumer, predicate Predicate) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.IsClosed() {
		return errors.New("producer closed")
	}

	for _, rec := range p.consumers {
		if rec.consumer == c {
			log.Panic("duplicate subscription")
		}
	}

	// Execute the callback asynchronously to prevent deadlock if it executes
	// immediately. This can only happen if the consumer is closed while
	// subscribing.
	if !c.OnClose(func() { go p.delete(c) }) {
		return errors.New("consumer closed")
	}
	p.consumers = append(p.consumers, subscription{consumer: c, predicate: predicate})

	// Put cached messages into consumer in a go routine because receiving on it
	// probably starts after subscription.
	cached := p.cache.Messages(predicate)
	go func() {
		for _, m := range cached {
			c.Put(m)
		}
	}()

	return nil
}

func (p *Relay) delete(c Consumer) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.IsClosed() {
		return
	}

	for i, sub := range p.consumers {
		if sub.consumer == c {
			p.consumers[i] = p.consumers[len(p.consumers)-1]
			p.consumers[len(p.consumers)-1] = subscription{} // For the GC.
			p.consumers = p.consumers[:len(p.consumers)-1]

			return
		}
	}
	log.Panic("deleted consumer that was not subscribed")
}

func (p *Relay) isEmpty() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return len(p.consumers) == 0
}

// Put puts an Envelope in the relay.
func (p *Relay) Put(e *Envelope) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.IsClosed() {
		return
	}

	any := false
	for _, sub := range p.consumers {
		if sub.predicate(e) {
			sub.consumer.Put(e)
			any = true
		}
	}

	if !any {
		if !p.cache.Put(e) {
			p.defaultMsgHandler(e)
		}
	}
}

func logUnhandledMsg(e *Envelope) {
	log.WithField("sender", e.Sender).
		WithField("recipient", e.Recipient).
		Debugf("Received %T message without subscription: %v", e.Msg, e.Msg)
}

// SetDefaultMsgHandler sets the default message handler.
func (p *Relay) SetDefaultMsgHandler(handler func(*Envelope)) {
	if handler == nil {
		handler = logUnhandledMsg
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.defaultMsgHandler = handler
}
