// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"context"
	stdsync "sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/sync"
)

// producer handles (un)registering Consumers for a message producer's messages.
type producer struct {
	sync.Closer
	mutex     stdsync.RWMutex
	consumers []subscription

	cache             Cache
	defaultMsgHandler func(Msg) // Handles messages with no subscriber.
}

type subscription struct {
	consumer  Consumer
	predicate Predicate
}

func (p *producer) Close() error {
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
func (p *producer) Cache(ctx context.Context, predicate Predicate) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.IsClosed() {
		return
	}

	p.cache.Cache(ctx, predicate)
}

// Subscribe adds a receiver to the subscriptions.
// If the receiver was already subscribed, Subscribe panics.
// If the peer is closed, Subscribe returns an error.
// Otherwise, Subscribe returns nil.
func (p *producer) Subscribe(c Consumer, predicate Predicate) error {
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
	cached := p.cache.Get(predicate)
	go func() {
		for _, m := range cached {
			c.Put(m.Annex.(*Endpoint), m.Msg)
		}
	}()

	return nil
}

func (p *producer) delete(c Consumer) {
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

func (p *producer) isEmpty() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return len(p.consumers) == 0
}

func (p *producer) produce(m Msg, peer *Endpoint) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.IsClosed() {
		return
	}

	any := false
	for _, sub := range p.consumers {
		if sub.predicate(m) {
			sub.consumer.Put(peer, m)
			any = true
		}
	}

	if !any {
		if !p.cache.Put(m, peer) {
			p.defaultMsgHandler(m)
		}
	}
}

func logUnhandledMsg(m Msg) {
	log.Debugf("Received %T message without subscription: %v", m, m)
}

func (p *producer) SetDefaultMsgHandler(handler func(Msg)) {
	if handler == nil {
		handler = logUnhandledMsg
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.defaultMsgHandler = handler
}

func makeProducer() producer {
	return producer{defaultMsgHandler: logUnhandledMsg}
}
