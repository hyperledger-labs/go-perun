// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	wire "perun.network/go-perun/wire/msg"
)

// subscriptions handles (un)registering Receivers for a peer's messages.
// It is separate from Peer to reduce the complexity of that type.
type subscriptions struct {
	mutex sync.RWMutex
	subs  []subscription
	peer  *Peer
}

type subscription struct {
	receiver  *Receiver
	predicate func(wire.Msg) bool
}

// add adds a receiver to the subscriptions.
// If the receiver was already subscribed, panics.
// If the peer is closed, returns an error.
func (s *subscriptions) add(predicate func(wire.Msg) bool, r *Receiver) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.peer.isClosed() {
		return errors.New("peer closed")
	}

	for _, rec := range s.subs {
		if rec.receiver == r {
			log.Panic("duplicate peer subscription")
		}
	}

	s.subs = append(s.subs, subscription{receiver: r, predicate: predicate})

	return nil
}

func (s *subscriptions) delete(r *Receiver) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, sub := range s.subs {
		if sub.receiver == r {
			s.subs[i] = s.subs[len(s.subs)-1]
			s.subs = s.subs[:len(s.subs)-1]

			return
		}
	}
	log.Panic("deleted receiver that was not subscribed")
}

func (s *subscriptions) isEmpty() bool {
	return len(s.subs) == 0
}

func (s *subscriptions) put(m wire.Msg, p *Peer) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, sub := range s.subs {
		if sub.predicate(m) {
			sub.receiver.msgs <- msgTuple{p, m}
		}
	}
}

func makeSubscriptions(p *Peer) subscriptions {
	return subscriptions{peer: p}
}
