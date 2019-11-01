// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	wire "perun.network/go-perun/wire/msg"
)

const (
	// receiverBufferSize controls how many messages can be queued in a
	// receiver before blocking.
	receiverBufferSize = 16
)

// msgTuple is a helper type, because channels cannot have tuple types.
type msgTuple struct {
	*Peer
	wire.Msg
}

// Receiver is a helper object that can subscribe to different message
// categories from multiple peers. Receivers must only be used by a single
// execution context at a time. If multiple contexts need to access a peer's
// messages, then multiple receivers have to be created.
//
// Receivers have two ways of accessing peer messages: Next() and NextWait().
// Both will return a channel to the next message, but the difference is that
// Next() will fail when the receiver is not subscribed to any peers, while
// NextWait() will only fail if the receiver is manually closed via Close().
type Receiver struct {
	mutex  sync.Mutex    // Protects all fields.
	msgs   chan msgTuple // Queued messages.
	closed bool
	subs   []*Peer // The receiver's subscription list.
}

// Subscribe subscribes a receiver to all of a peer's messages of the requested
// message category. Returns an error if the receiver is closed.
func (r *Receiver) Subscribe(p *Peer, predicate func(wire.Msg) bool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.closed {
		return errors.New("receiver is closed")
	}

	if err := p.subs.add(predicate, r); err != nil {
		return err
	}

	r.subs = append(r.subs, p)

	return nil
}

// Unsubscribe removes a receiver's subscription to a peer's messages of the
// requested category. Returns an error if the receiver was not subscribed to
// the requested peer and message category.
func (r *Receiver) Unsubscribe(p *Peer) {
	r.unsubscribe(p, true)
}

func (r *Receiver) unsubscribe(p *Peer, doDelete bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for i, _p := range r.subs {
		if _p == p {
			if doDelete {
				p.subs.delete(r)
			}
			r.subs[i] = r.subs[len(r.subs)-1]
			r.subs = r.subs[:len(r.subs)-1]

			return
		}
	}
	log.Panic("unsubscribe called on not-subscribed source")
}

// UnsubscribeAll removes all of a receiver's subscriptions.
func (r *Receiver) UnsubscribeAll() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.unsubscribeAll()
}

func (r *Receiver) unsubscribeAll() {
	for _, p := range r.subs {
		p.subs.delete(r)
	}
	r.subs = nil
}

// Next returns a channel to the next message.
func (r *Receiver) Next(ctx context.Context) (*Peer, wire.Msg) {
	select {
	case <-ctx.Done():
		return nil, nil
	case tuple := <-r.msgs:
		return tuple.Peer, tuple.Msg
	}
}

// Close closes a receiver.
// Any ongoing receiver operations will be aborted (if there are no messages in
// backlog).
func (r *Receiver) Close() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.closed {
		r.closed = true
		// Remove all subscriptions.
		r.unsubscribeAll()
		// Close the message channel.
		close(r.msgs)
	}
}

// NewReceiver creates a new receiver.
func NewReceiver() *Receiver {
	return &Receiver{
		msgs: make(chan msgTuple, receiverBufferSize),
	}
}
