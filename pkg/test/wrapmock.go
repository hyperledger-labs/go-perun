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
	"runtime"
	"strings"
	"testing"
)

type (
	// WrapMock is a mocking object to test whether an object's methods are
	// called by an outer function with the same name, thus being wrapped.
	// This is particularly useful for global objects like the global logger that
	// have wrapped package function calls.
	WrapMock struct {
		t      testingErrorer
		called bool
	}

	// testingErrorer is a testing type on which we can call Error()/Errorf().
	// This abstraction is needed in order to be able to test WrapMock and
	// check that the Error functions have been called.
	testingErrorer interface {
		Error(...interface{})
		Errorf(string, ...interface{})
	}
)

// NewWrapMock creates a new mock for wrapped objects.
func NewWrapMock(t *testing.T) WrapMock {
	return WrapMock{t: t}
}

// AssertWrapped asserts that the two function names in the stack above
// AssertWrapped() are the same. This means that the mocked object's method
// calling AssertWrapped() is wrapped.
// The fact that the method was called is also recorded and can be asserted with
// AssertCalled().
//
// All method implementations of the object should just call this method.
func (w *WrapMock) AssertWrapped() {
	w.called = true
	// record two next outer frames
	pc := make([]uintptr, 2)
	// skip inner two frames "AssertWrapped", and "runtime.Callers"
	runtime.Callers(2, pc)

	frames := runtime.CallersFrames(pc)
	methodFrame, _ := frames.Next()
	packageFrame, _ := frames.Next()
	// Frame.Function has the form "perun.network/path/to/pkg.(*Type).fn" for
	// method "fn" on object "Type" and the form "perun.network/path/to/pkg.fn"
	// for package functions
	methodFn := splitDotLast(methodFrame.Function)
	packageFn := splitDotLast(packageFrame.Function)

	// check matching function names
	if methodFn != packageFn {
		w.t.Errorf("called method %q not wrapped, called by %q instead.", methodFn, packageFn)
	}
}

// AssertCalled asserts that the object was called and resets the called flag
// afterwards.
func (w *WrapMock) AssertCalled() {
	if !w.called {
		w.t.Error("object was not called")
	}
	w.called = false
}

func splitDotLast(s string) string {
	ss := strings.Split(s, ".")
	return ss[len(ss)-1]
}
