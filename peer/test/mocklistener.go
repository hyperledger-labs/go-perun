// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/peer/test"

import (
	"sync/atomic"

	"github.com/pkg/errors"

	"perun.network/go-perun/peer"
)

var _ peer.Listener = (*MockListener)(nil)

// MockListener is a mocked listener that can be used to control and examine a
// listener. Accept() calls can be manually controlled via Put(). Accepted()
// tracks the number of accepted connections. IsClosed() can be used to detect
// whether a MockListener is still open.
type MockListener struct {
	closed chan struct{}  // Whether the listener is closed.
	queue  chan peer.Conn // The connection queue (unbuffered).

	accepted int32 // The number of connections that have been accepted.
}

// NewMockListener creates a new mock listener.
func NewMockListener() *MockListener {
	return &MockListener{
		closed:   make(chan struct{}),
		queue:    make(chan peer.Conn),
		accepted: 0,
	}
}

// Accept returns the next connection that is enqueued via Put(). This function
// blocks until either Put() is called or until the listener is closed.
func (m *MockListener) Accept() (peer.Conn, error) {
	if m.IsClosed() {
		return nil, errors.New("listener closed")
	}

	select {
	case <-m.closed:
		return nil, errors.New("listener closed")
	case conn := <-m.queue:
		atomic.AddInt32(&m.accepted, 1)
		return conn, nil
	}
}

// Close closes the mock listener.
// This aborts any ongoing Accept() call and all future Accept() calls will
// fail. If the listener is already closed, returns an error.
func (m *MockListener) Close() (err error) {
	defer func() { recover() }()
	err = errors.New("listener already closed")
	close(m.closed)
	err = nil
	return
}

// IsClosed returns whether the listener is closed.
func (m *MockListener) IsClosed() bool {
	select {
	case <-m.closed:
		return true
	default:
		return false
	}
}

// Put enqueues one connection to be returned by Accept().
// If the listener is already closed, does nothing. This function blocks until
// either Accept() is called or until the listener is closed.
//
// Note that if Put() is called in parallel, there is no ordering guarantee for
// the accepted connections.
func (m *MockListener) Put(conn peer.Conn) {
	select {
	case m.queue <- conn:
	case <-m.closed:
		return
	}
}

// NumAccepted returns the number of connections that have been accepted by the
// listener. Note that this number is updated before Accept() returns, but not
// necessarily before Put() returns.
func (m *MockListener) NumAccepted() int {
	return int(atomic.LoadInt32(&m.accepted))
}
