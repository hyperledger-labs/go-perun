// Copyright 2024 - See NOTICE file for copyright holders.
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

package test

import (
	"context"
	"net"
	"sync/atomic"

	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"

	"perun.network/go-perun/wire"
	wirenet "perun.network/go-perun/wire/net"
	"polycry.pt/poly-go/sync"
)

// Dialer is a test dialer that can dial connections to Listeners via a ConnHub.
type Dialer struct {
	hub    *ConnHub
	dialed int32

	sync.Closer
}

var _ wirenet.Dialer = (*Dialer)(nil)

// NewDialer creates a new test dialer.
func NewDialer(hub *ConnHub) *Dialer {
	return &Dialer{
		hub: hub,
	}
}

// Dial tries to connect to a wire.
func (d *Dialer) Dial(ctx context.Context, address map[wallet.BackendID]wire.Address, ser wire.EnvelopeSerializer) (wirenet.Conn, error) {
	if d.IsClosed() {
		return nil, errors.New("dialer closed")
	}

	select {
	case <-ctx.Done():
		return nil, errors.New("manually aborted")
	default:
	}

	l, ok := d.hub.find(address)
	if !ok {
		return nil, errors.Errorf("peer with address %v not found", address)
	}

	local, remote := net.Pipe()
	if !l.Put(ctx, wirenet.NewIoConn(remote, ser)) {
		local.Close()
		remote.Close()
		return nil, errors.New("Put() failed")
	}
	atomic.AddInt32(&d.dialed, 1)
	return wirenet.NewIoConn(local, ser), nil
}

// Close closes a connection.
func (d *Dialer) Close() error {
	return errors.WithMessage(d.Closer.Close(), "dialer was already closed")
}

// NumDialed returns how many peers have been dialed.
func (d *Dialer) NumDialed() int {
	return int(atomic.LoadInt32(&d.dialed))
}
