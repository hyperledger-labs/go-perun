// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/sync"
)

const (
	// receiverBufferSize controls how many messages can be queued in a
	// receiver before blocking.
	receiverBufferSize = 16
)

var _ Consumer = (*Receiver)(nil)

// Receiver is a helper object that can subscribe to different message
// categories from multiple peers. Receivers must only be used by a single
// execution context at a time. If multiple contexts need to access a peer's
// messages, then multiple receivers have to be created.
type Receiver struct {
	msgs chan *Envelope

	sync.Closer
}

// Next returns a channel to the next message.
func (r *Receiver) Next(ctx context.Context) (*Envelope, error) {
	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context closed")
	case <-r.Closed():
		return nil, errors.New("receiver closed")
	default:
	}

	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context closed")
	case <-r.Closed():
		return nil, errors.New("receiver closed")
	case e := <-r.msgs:
		return e, nil
	}
}

// Put puts a new message into the queue.
func (r *Receiver) Put(e *Envelope) {
	select {
	case r.msgs <- e:
	case <-r.Closed():
	}
}

// NewReceiver creates a new receiver.
func NewReceiver() *Receiver {
	return &Receiver{
		msgs: make(chan *Envelope, receiverBufferSize),
	}
}
