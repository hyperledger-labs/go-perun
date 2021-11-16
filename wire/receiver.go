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

	"github.com/pkg/errors"

	"polycry.pt/poly-go/sync"
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
