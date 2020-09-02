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
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
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
	errs  *accumulatedError
	wg    sync.WaitGroup

	failed chan struct{} // Closed when an error has occurred.
}

// Failed returns a channel that is closed when an error occurs.
func (g *Gatherer) Failed() <-chan struct{} {
	return g.failed
}

// Add gathers an error. If the supplied error is nil, it is ignored. On the
// first error, closes the channel returned by Failed.
func (g *Gatherer) Add(err error) {
	if err == nil {
		return
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.errs == nil {
		g.errs = &accumulatedError{}
	}

	g.errs.errs = append(g.errs.errs, err)

	select {
	case <-g.failed:
	default:
		close(g.failed)
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
// returned.
func (g *Gatherer) Wait() {
	g.wg.Wait()
}

// Err returns the accumulated error. If there are no errors, returns nil.
func (g *Gatherer) Err() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if g.errs == nil { // Because (*accumulatedError)(nil) != (error)(nil).
		return nil
	}
	return g.errs
}

// stackTracer is taken from the github.com/pkg/errors documentation.
type stackTracer interface {
	StackTrace() errors.StackTrace
}

type accumulatedError struct {
	errs []error
}

// Error returns an error message containing all the sub-errors that occurred.
func (e *accumulatedError) Error() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("(%d error", len(e.errs)))
	if len(e.errs) != 1 {
		builder.WriteByte('s')
	}
	builder.WriteByte(')')
	for i, err := range e.errs {
		builder.WriteString(fmt.Sprintf("\n%d): %s", i+1, err.Error()))
	}
	return builder.String()
}

// StackTrace returns the first available stack trace, or none.
func (e *accumulatedError) StackTrace() errors.StackTrace {
	for _, err := range e.errs {
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
	if acc, ok := cerr.(*accumulatedError); ok {
		return acc.errs
	}
	return []error{err}
}
