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

package test

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// Abort describes the reason for an extraordinary function termination. It is
// either of type Panic or Goexit, or nil.
type Abort interface {
	// Stack returns the stack trace of the termination's cause.
	Stack() string
	// String returns a textual representation of the Abort cause.
	String() string
}

type abortBase struct {
	stack string
}

func (a abortBase) Stack() string {
	return a.stack
}

// Panic describes a recovered runtime.Goexit() or panic(), containing the
// original message (in case of a panic) and the stack trace that caused the panic().
type Panic struct {
	abortBase
	value interface{} // The argument to panic() that was used.
}

// String formats the abort so that it can be printed similar to the native
// panic printing.
func (p Panic) String() string {
	return fmt.Sprintf("panic: %v\n\n%s", p.value, p.stack)
}

// Value returns the value that was passed to Panic().
func (p Panic) Value() interface{} {
	return p.value
}

// Goexit describes a recovered runtime.Goexit().
type Goexit struct {
	abortBase
}

func (g Goexit) String() string {
	return "runtime.Goexit:\n\n" + g.Stack()
}

// CheckAbort tests whether a supplied function is aborted early using panic()
// or runtime.Goexit(). Returns a descriptor of the termination cause or nil if
// it terminated normally.
func CheckAbort(function func()) (abort Abort) {
	done := make(chan struct{})

	goexit := true  // Whether runtime.Goexit occurred.
	aborted := true // Whether panic or runtime.Goexit occurred.

	var base abortBase        // The abort cause's stack trace.
	var recovered interface{} // The recovered panic() value.

	go func() {
		defer close(done)
		func() {
			defer func() {
				if aborted {
					// Recover the panic's value, and if it was a Panic already,
					// do not wrap it again.
					recovered = recover()
					if p, ok := recovered.(*Panic); ok {
						base.stack = p.Stack()
						recovered = p.Value()
					} else {
						// Hide all mentions of CheckAbort and its inner
						// functions as well as panic and Goexit.
						base.stack = getStack(false, 2, 3)
					}
				}
			}()
			// Call the function to be checked.
			function()
			aborted = false
		}()

		// This is executed if Goexit was not called.
		goexit = false
	}()

	<-done

	// Concatenate the inner call stack of the failure (which starts at the
	// goroutine instantiation) with the goroutine that is calling CheckAbort.
	if goexit || aborted {
		base.stack += "\n" + getStack(true, 1, 0)
	}

	if goexit {
		abort = &Goexit{base}
	} else if aborted {
		abort = &Panic{base, recovered}
	}
	return
}

// getStack retrieves the current call stack as text, and optionally removes the
// first line ("goroutine XXX [running]:") and an optional number of the
// inner-most and outer-most stack frames.
func getStack(hideGoroutine bool, hideInnerCallers, hideOuterCallers int) string {
	goroutine, stack := removeLine(string(debug.Stack()))

	// getStack() + debug.Stack() + hideInnerCallers.
	removeInnerFunctions := 2 + hideInnerCallers
	for i := 0; i < 2*removeInnerFunctions; i++ {
		_, stack = removeLine(stack)
	}

	for i := 0; i < 2*hideOuterCallers; i++ {
		stack, _ = removeLastLine(stack)
	}

	if !hideGoroutine {
		stack = goroutine + "\n" + stack
	}
	return stack
}

func removeLine(str string) (line, rest string) {
	i := strings.Index(str, "\n")
	if i == -1 {
		i = len(str)
	}
	return str[:i], str[i+1:]
}

func removeLastLine(str string) (start, line string) {
	str = strings.TrimSuffix(str, "\n") // Ignore newline at end of string.
	i := strings.LastIndex(str, "\n")
	if i == -1 {
		return str, ""
	}
	return str[:i], str[i+1:]
}

// CheckGoexit tests whether a supplied function calls runtime.Goexit during its
// execution. Rethrows panics, but wrapped into a Panic object to preserve the
// stack trace and value passed to panic().
// Returns whether the supplied function did call runtime.Goexit.
func CheckGoexit(function func()) bool {
	abort := CheckAbort(function)
	if p, ok := abort.(*Panic); ok {
		panic(p)
	}
	return abort != nil
}
