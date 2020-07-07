// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"context"

	"perun.network/go-perun/pkg/sync"
)

const (
	// receiverBufferSize controls how many messages can be queued in a
	// receiver before blocking.
	receiverBufferSize = 16
)

// msgTuple is a helper type, because channels cannot have tuple types.
type msgTuple struct {
	*Endpoint
	Msg
}

var _ Consumer = (*Receiver)(nil)

// Receiver is a helper object that can subscribe to different message
// categories from multiple peers. Receivers must only be used by a single
// execution context at a time. If multiple contexts need to access a peer's
// messages, then multiple receivers have to be created.
type Receiver struct {
	msgs chan msgTuple // Queued messages.

	sync.Closer
}

// Next returns a channel to the next message.
func (r *Receiver) Next(ctx context.Context) (*Endpoint, Msg) {
	select {
	case <-ctx.Done():
		return nil, nil
	case <-r.Closed():
		return nil, nil
	default:
	}

	select {
	case <-ctx.Done():
		return nil, nil
	case <-r.Closed():
		return nil, nil
	case tuple := <-r.msgs:
		return tuple.Endpoint, tuple.Msg
	}
}

// Put puts a new message into the queue.
func (r *Receiver) Put(peer *Endpoint, msg Msg) {
	select {
	case r.msgs <- msgTuple{peer, msg}:
	case <-r.Closed():
	}
}

// NewReceiver creates a new receiver.
func NewReceiver() *Receiver {
	return &Receiver{
		msgs: make(chan msgTuple, receiverBufferSize),
	}
}
