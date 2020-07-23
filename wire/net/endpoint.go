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

package net

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/wire"
)

// Endpoint is an authenticated connection to a Perun node.
// It contains the node's identity. Endpoints are thread-safe.
// Endpoints must not be created manually. The creation of Endpoints is handled
// by the Registry, which tracks all existing Endpoints. The registry, in turn,
// is used by the Bus.
//
// Sending messages to a node is done via the Send() method. To receive messages
// from an Endpoint, use the Receiver helper type (by subscribing).
type Endpoint struct {
	Address wire.Address // The Endpoint's Perun address.
	conn    Conn         // The Endpoint's connection.

	sending sync.Mutex // Blocks multiple Send calls.
}

// recvLoop continuously receives messages from an Endpoint until it is closed.
// Received messages are relayed via the Endpoint's subscription system. This is
// called by the registry when the Endpoint is registered.
func (p *Endpoint) recvLoop(c wire.Consumer) {
	for {
		e, err := p.conn.Recv()
		if err != nil {
			p.Close() // Ignore double close.
			log.WithError(err).Errorf("Ending recvLoop on closed connection of Endpoint %v", p.Address)
			return
		}
		// Emit the received envelope.
		c.Put(e)
	}
}

// Send sends a single message to an Endpoint.
// Fails if the Endpoint is closed via Close() or the transmission fails.
//
// The passed context is used to timeout the send operation. If the context
// times out, the Endpoint is closed.
func (p *Endpoint) Send(ctx context.Context, e *wire.Envelope) error {
	if !p.sending.TryLockCtx(ctx) {
		p.Close()
		return errors.New("failed to lock sending mutex")
	}

	sent := make(chan error, 1)
	// Asynchronously send, because we cannot abort Conn.Send().
	go func() {
		defer p.sending.Unlock()
		sent <- p.conn.Send(e)
	}()

	// Return as soon as the sending finishes, times out, or Endpoint is closed.
	select {
	case err := <-sent:
		return err
	case <-ctx.Done():
		p.Close()
		return errors.Wrap(ctx.Err(), "context canceled")
	}
}

// Close closes the Endpoint's connection. A closed Endpoint is no longer usable.
func (p *Endpoint) Close() (err error) {
	return p.conn.Close()
}

// newEndpoint creates a new Endpoint from a wire Address and connection.
func newEndpoint(addr wire.Address, conn Conn) *Endpoint {
	return &Endpoint{
		Address: addr,
		conn:    conn,
	}
}

// String returns the Endpoint's address string.
func (p *Endpoint) String() string {
	return p.Address.String()
}
