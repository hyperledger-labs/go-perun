// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	stdsync "sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/sync"
	wire "perun.network/go-perun/wire/msg"
)

// producer handles (un)registering Consumers for a message producer's messages.
type producer struct {
	mutex     stdsync.RWMutex
	consumers []subscription
	sync.Closer

	defaultMsgHandler func(wire.Msg) // Handles messages with no subscriber.
}

type subscription struct {
	consumer  Consumer
	predicate func(wire.Msg) bool
}

func (p *producer) Close() error {
	if err := p.Closer.Close(); err != nil {
		return err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.consumers = nil
	return nil
}

// add adds a receiver to the subscriptions.
// If the receiver was already subscribed, panics.
// If the peer is closed, returns an error.
func (p *producer) Subscribe(c Consumer, predicate func(wire.Msg) bool) error {
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
	log.Panic("deleted receiver that was not subscribed")
}

func (p *producer) isEmpty() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return len(p.consumers) == 0
}

func (p *producer) produce(m wire.Msg, peer *Peer) {
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
		p.defaultMsgHandler(m)
	}
}

func logUnhandledMsg(m wire.Msg) {
	log.Debugf("Received %T message without subscription: %v", m, m)
}

func (p *producer) SetDefaultMsgHandler(handler func(wire.Msg)) {
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
