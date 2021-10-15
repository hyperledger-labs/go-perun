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

package sync

import (
	"context"
	"sync"
	"time"

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

// OnClose registers the passed callback to be executed when the Closer is
// closed. If the closer is already closed, the callback will be ignored.
//
// Returns true if the callback was successfully registered and false if the
// closer was already closed and callback was ignored.
func (c *Closer) OnClose(handler func()) bool {
	c.onClosedMtx.Lock()
	defer c.onClosedMtx.Unlock()
	// Check again, because Close might have been called before the lock was
	// acquired.
	if c.IsClosed() {
		return false
	}
	c.onClosed = append(c.onClosed, handler)
	return true
}

// OnCloseAlways registers the passed callback to be executed when the Closer
// is closed. If the closer is already closed, the callback will be executed
// immediately.
//
// Returns true if the callback was successfully registered and false if the
// closer was already closed and callback was executed immediately.
func (c *Closer) OnCloseAlways(handler func()) bool {
	c.onClosedMtx.Lock()
	defer c.onClosedMtx.Unlock()
	// Check again, because Close might have been called before the lock was
	// acquired.
	if c.IsClosed() {
		handler()
		return false
	}
	c.onClosed = append(c.onClosed, handler)
	return true
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

// implementation of a Closer as a context.Context.
type closerCtx Closer

// Ctx returns a context that is canceled when the Closer is closed.
func (c *Closer) Ctx() context.Context { return (*closerCtx)(c) }

// Deadline implements context.Deadline trivially - it returns 0, false.
func (c *closerCtx) Deadline() (deadline time.Time, ok bool) { return }

// Done is closed when the Closer is closed.
func (c *closerCtx) Done() <-chan struct{} { return (*Closer)(c).Closed() }

// If the Closer is not yet closed, Err returns nil.
// If the Closer is closed, Err returns a context-canceled error.
func (c *closerCtx) Err() error {
	if (*Closer)(c).IsClosed() {
		return context.Canceled
	}
	return nil
}

// Value always returns nil. It is just there to implement the context.Context
// interface.
func (c *closerCtx) Value(interface{}) interface{} { return nil }
