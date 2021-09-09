// Copyright 2020 - See NOTICE file for copyright holders.
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

package errors

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"

	pkgsync "perun.network/go-perun/pkg/sync"
)

// NewGatherer creates a new error gatherer.
func NewGatherer() *Gatherer {
	return &Gatherer{failed: make(chan struct{})}
}

// Gatherer accumulates errors into a single error, similar to an error group,
// but retains all occurred errors instead of just the first one. Use
// NewGatherer() to create error gatherers.
type Gatherer struct {
	mutex sync.Mutex
	errs  accumulatedError
	wg    pkgsync.WaitGroup

	onFails []func()
	failed  chan struct{} // Closed when an error has occurred.
}

// Failed returns a channel that is closed when an error occurs.
func (g *Gatherer) Failed() <-chan struct{} {
	return g.failed
}

// WaitDoneOrFailed returns when either:
// all routines returned successfully OR
// any routine returned with an error.
func (g *Gatherer) WaitDoneOrFailed() {
	_ = g.WaitDoneOrFailedCtx(context.Background())
}

// WaitDoneOrFailedCtx returns when either:
// all routines returned successfully OR
// any routine returned with an error OR
// the context was cancelled.
// The result indicates whether the context was cancelled.
func (g *Gatherer) WaitDoneOrFailedCtx(ctx context.Context) bool {
	select {
	case <-g.failed:
	case <-g.wg.WaitCh():
	case <-ctx.Done():
	}
	return ctx.Err() == nil
}

// Add gathers an error. If the supplied error is nil, it is ignored. On the
// first error, closes the channel returned by Failed.
func (g *Gatherer) Add(err error) {
	if err == nil {
		return
	}

	g.mutex.Lock()
	g.errs = append(g.errs, err)
	g.mutex.Unlock()

	select {
	case <-g.failed:
		return
	default:
	}

	close(g.failed)

	for _, fn := range g.onFails {
		fn()
	}
}

// Go executes a function in a goroutine and gathers its returned error.
func (g *Gatherer) Go(fn func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		g.Add(fn())
	}()
}

// Wait waits until all goroutines that have been launched via Go() have
// returned and returns the accumulated error.
func (g *Gatherer) Wait() error {
	g.wg.Wait()
	return g.Err()
}

// Err returns the accumulated error. If there are no errors, returns nil.
func (g *Gatherer) Err() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if g.errs == nil { // Because accumulatedError(nil) != error(nil).
		return nil
	}
	return g.errs
}

// OnFail adds fn to the list of functions that are executed right after any
// non-nil error is added with Add (or any routine started with Go failed). The
// functions are guaranteed to be executed in the order that they were added.
//
// The channel returned by Failed is closed before those functions are executed.
func (g *Gatherer) OnFail(fn func()) {
	g.onFails = append(g.onFails, fn)
}

// stackTracer is taken from the github.com/pkg/errors documentation.
type stackTracer interface {
	StackTrace() errors.StackTrace
}

type accumulatedError []error

// Error returns an error message containing all the sub-errors that occurred.
func (e accumulatedError) Error() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("(%d error", len(e)))
	if len(e) != 1 {
		builder.WriteByte('s')
	}
	builder.WriteByte(')')
	for i, err := range e {
		builder.WriteString(fmt.Sprintf("\n%d): %s", i+1, err.Error()))
	}
	return builder.String()
}

// StackTrace returns the first available stack trace, or none.
func (e accumulatedError) StackTrace() errors.StackTrace {
	for _, err := range e {
		if s, ok := err.(stackTracer); ok {
			return s.StackTrace()
		}
	}
	return nil
}

// Causes returns an error's causes as a slice. If an error is not a compound
// error, returns a slice containing only the passed error. Returns a nil slice
// for nil errors.
func Causes(err error) []error {
	if err == nil {
		return nil
	}

	cerr := errors.Cause(err)
	if acc, ok := cerr.(accumulatedError); ok {
		return acc
	}
	return []error{err}
}
