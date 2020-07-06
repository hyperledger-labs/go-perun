// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"context"
	"sync/atomic"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/wire"
)

var _ wire.Listener = &Listener{}

// Listener is a mocked listener that can be used to control and examine a
// listener. Accept() calls can be manually controlled via Put(). Accepted()
// tracks the number of accepted connections. IsClosed() can be used to detect
// whether a Listener is still open.
type Listener struct {
	sync.Closer
	queue chan wire.Conn // The connection queue (unbuffered).

	accepted int32 // The number of connections that have been accepted.
}

// NewListener creates a new test listener.
func NewListener() *Listener {
	return &Listener{
		queue:    make(chan wire.Conn),
		accepted: 0,
	}
}

// Accept returns the next connection that is enqueued via Put(). This function
// blocks until either Put() is called or until the listener is closed.
func (l *Listener) Accept() (wire.Conn, error) {
	if l.IsClosed() {
		return nil, errors.New("listener closed")
	}

	select {
	case <-l.Closed():
		return nil, errors.New("listener closed")
	case conn := <-l.queue:
		atomic.AddInt32(&l.accepted, 1)
		return conn, nil
	}
}

// Close closes the test listener.
// This aborts any ongoing Accept() call and all future Accept() calls will
// fail. If the listener is already closed, returns an error.
func (l *Listener) Close() error {
	return errors.WithMessage(l.Closer.Close(), "listener was already closed")
}

// Put enqueues one connection to be returned by Accept().
// If the listener is already closed, does nothing. This function blocks until
// either Accept() is called or until the listener is closed. Returns whether
// the connection has been retrieved by Accept().
//
// Note that if Put() is called in parallel, there is no ordering guarantee for
// the accepted connections.
func (l *Listener) Put(ctx context.Context, conn wire.Conn) bool {
	select {
	case l.queue <- conn:
		return true
	case <-l.Closed():
		return false
	case <-ctx.Done():
		return false
	}
}

// NumAccepted returns the number of connections that have been accepted by the
// listener. Note that this number is updated before Accept() returns, but not
// necessarily before Put() returns.
func (l *Listener) NumAccepted() int {
	return int(atomic.LoadInt32(&l.accepted))
}
