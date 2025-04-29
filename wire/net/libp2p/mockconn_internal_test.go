// Copyright 2025 - See NOTICE file for copyright holders.
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

package libp2p

import (
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/wire"
	wirenet "perun.network/go-perun/wire/net"
	"polycry.pt/poly-go/sync/atomic"
)

var _ wirenet.Conn = (*MockConn)(nil)

type MockConn struct {
	mutex     sync.Mutex
	closed    atomic.Bool
	recvQueue chan *wire.Envelope

	sent func(*wire.Envelope) // observes sent messages.
}

func newMockConn() *MockConn {
	return &MockConn{
		sent:      func(*wire.Envelope) {},
		recvQueue: make(chan *wire.Envelope, 1),
	}
}

func (c *MockConn) Send(e *wire.Envelope) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed.IsSet() {
		return errors.New("closed")
	}
	c.sent(e)
	return nil
}

func (c *MockConn) Recv() (*wire.Envelope, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed.IsSet() {
		return nil, errors.New("closed")
	}
	return <-c.recvQueue, nil
}

func (c *MockConn) Close() error {
	if !c.closed.TrySet() {
		return errors.New("double close")
	}
	return nil
}
