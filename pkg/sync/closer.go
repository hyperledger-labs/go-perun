// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package sync

import (
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/sync/atomic"
)

// Closer is a utility type for implementing an "onclose" event.
// It supports registering handlers, waiting for the event, and status checking.
// A default-initialised Closer is a valid value.
type Closer struct {
	once     sync.Once     // Initializes the closed channel.
	isClosed atomic.Bool   // Whether the Closer is currently closed.
	closed   chan struct{} // Closed when the Closer is closed.

	onClosedMtx sync.Mutex // Protects callbacks.
	onClosed    []func()   // Executed when Close() is called.
}

// OnCloser contains the OnClose and OnCloseAlways function.
type OnCloser interface {
	OnClose(func()) bool
	OnCloseAlways(func()) bool
}

func (c *Closer) initOnce() {
	c.once.Do(func() { c.closed = make(chan struct{}) })
}

// Closed returns a channel to be used in a select statement.
func (c *Closer) Closed() <-chan struct{} {
	c.initOnce()
	return c.closed
}

// Close executes all registered callbacks.
// If Close was already called before, returns an AlreadyClosedError, otherwise,
// returns nil.
func (c *Closer) Close() error {
	c.initOnce()

	if !c.isClosed.TrySet() {
		return newAlreadyClosedError()
	}

	c.onClosedMtx.Lock()
	defer c.onClosedMtx.Unlock()
	for _, fn := range c.onClosed {
		fn()
	}

	close(c.closed)

	return nil
}

// IsClosed returns whether the Closer is currently closed.
func (c *Closer) IsClosed() bool {
	return c.isClosed.IsSet()
}

// OnClose registers the passed callback to be calle when the Closer is closed.
// If the Closer is already closed, does nothing. Returns whether the Closer was
// not yet closed.
func (c *Closer) OnClose(handler func()) bool {
	c.onClosedMtx.Lock()
	defer c.onClosedMtx.Unlock()
	// Check again, because Close might have been called before the lock was
	// acquired.
	if c.IsClosed() {
		return false
	} else {
		c.onClosed = append(c.onClosed, handler)
		return true
	}
}

// OnCloseAlways registers the passed callback to be called when the Closer is
// closed.
// If the Closer is already closed, immediately executes the callback. Returns
// whether the closer was not yet closed.
func (c *Closer) OnCloseAlways(handler func()) bool {
	c.onClosedMtx.Lock()
	defer c.onClosedMtx.Unlock()
	// Check again, because Close might have been called before the lock was
	// acquired.
	if c.IsClosed() {
		handler()
		return false
	} else {
		c.onClosed = append(c.onClosed, handler)
		return true
	}
}

var _ error = alreadyClosedError{}

type alreadyClosedError struct{}

const alreadyClosedMsg = "Closer already closed"

func (alreadyClosedError) Error() string {
	return alreadyClosedMsg
}

func newAlreadyClosedError() error {
	return errors.WithStack(alreadyClosedError{})
}

// IsAlreadyClosedError checks whether an error is an AlreadyClosedError.
func IsAlreadyClosedError(err error) bool {
	_, ok := errors.Cause(err).(alreadyClosedError)
	return ok
}
