// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"context"
	stdsync "sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/wire/msg"
)

// producer handles (un)registering Consumers for a message producer's messages.
type producer struct {
	sync.Closer
	mutex     stdsync.RWMutex
	consumers []subscription

	cache             msg.Cache
	defaultMsgHandler func(msg.Msg) // Handles messages with no subscriber.
}

type subscription struct {
	consumer  Consumer
	predicate msg.Predicate
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
func (p *producer) Cache(ctx context.Context, predicate msg.Predicate) {
	p.cache.Cache(ctx, predicate)
}

// add adds a receiver to the subscriptions.
// If the receiver was already subscribed, panics.
// If the peer is closed, returns an error.
func (p *producer) Subscribe(c Consumer, predicate msg.Predicate) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.IsClosed() {
		return errors.New("peer closed")
	}

	for _, rec := range p.consumers {
		if rec.consumer == c {
			log.Panic("duplicate peer subscription")
		}
	}

	p.consumers = append(p.consumers, subscription{consumer: c, predicate: predicate})
	c.OnClose(func() { p.delete(c) })

	// Put cached messages into consumer in a go routine because receiving on it
	// probably starts after subscription.
	cached := p.cache.Get(predicate)
	go func() {
		for _, m := range cached {
			c.Put(m.Annex.(*Peer), m.Msg)
		}
	}()

	return nil
}

func (p *producer) delete(c Consumer) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

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

func (p *producer) produce(m msg.Msg, peer *Peer) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

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

func logUnhandledMsg(m msg.Msg) {
	log.Debugf("Received %T message without subscription: %v", m, m)
}

func (p *producer) SetDefaultMsgHandler(handler func(msg.Msg)) {
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
