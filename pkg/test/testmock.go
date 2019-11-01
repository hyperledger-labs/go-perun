// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

// T is part of the interface that testing.T implements. Receive this type
// instead of *testing.T if you want to test your tests with the Tester.
type T interface {
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
}

// Tester is a testing.T mock to test tests.
// Create new instances of it with NewTester(t), passing it the actual
// *testing.T that the mock should call if an assertion fails.
// Then let the tests you want to test receive test.T instead of *testing.T and
// call your tests inside Tester.AssertX() calls.
type Tester struct {
	// testing object that should be used to report failures of tests. Usually,
	// this should be a *testing.T
	T
	fatalCalled   bool
	numErrorCalls uint
}

// NewTester creates a new testing mock, wrapping the passed actual test and
// calling t.Error() on it if a test test fails.
func NewTester(t T) *Tester {
	return &Tester{T: t}
}

func (t *Tester) Error(...interface{}) {
	t.err()
}

func (t *Tester) Errorf(string, ...interface{}) {
	t.err()
}

func (t *Tester) err() {
	t.numErrorCalls++
}

func (t *Tester) Fatal(...interface{}) {
	t.fatal()
}

func (t *Tester) Fatalf(string, ...interface{}) {
	t.fatal()
}

func (t *Tester) fatal() {
	t.fatalCalled = true
	// panic() to stop execution in the test, as would be the case in a test where
	// Fatal is called. The panic is recovered in the Assert... methods.
	panic("Tester.fatal()")
}

// AssertFatal checks that the passed function fn calls T.Fatal() on the T
// object it calls fn with.
func (t *Tester) AssertFatal(fn func(T)) {
	t.assert(func(t *Tester) {
		if !t.fatalCalled {
			t.T.Error("the test did not call Fatal()")
		}
	}, fn)
}

// AssertError checks that the passed function fn calls T.Error() on the T
// object it calls fn with.
func (t *Tester) AssertError(fn func(T)) {
	t.assert(func(t *Tester) {
		if t.numErrorCalls == 0 {
			t.T.Error("the test did not call Error()")
		}
	}, fn)
}

// AssertErrorN checks that the passed function fn calls T.Error() numCalls
// times on the T object it calls fn with.
func (t *Tester) AssertErrorN(fn func(T), numCalls uint) {
	t.assert(func(t *Tester) {
		if t.numErrorCalls != numCalls {
			t.T.Errorf("the test called Error() %d times, expected %d", t.numErrorCalls, numCalls)
		}
	}, fn)
}

// AssertErrorFatal checks that the passed function fn calls T.Error() and
// T.Fatal() on the T object it calls fn with.
func (t *Tester) AssertErrorFatal(fn func(T)) {
	t.assert(func(t *Tester) {
		if t.numErrorCalls == 0 && !t.fatalCalled {
			t.T.Error("the test did neither call Error() nor Fatal()")
		} else if t.numErrorCalls == 0 {
			t.T.Error("the test did not call Error() but Fatal()")
		} else if !t.fatalCalled {
			t.T.Error("the test did not call Fatal() but Error()")
		}
	}, fn)
}

// AssertErrorNFatal checks that the passed function fn calls T.Error() numCalls
// times and T.Fatal() on the T object it calls fn with.
func (t *Tester) AssertErrorNFatal(fn func(T), numCalls uint) {
	t.assert(func(t *Tester) {
		if t.numErrorCalls != numCalls && !t.fatalCalled {
			t.T.Errorf(
				"the test called Error() %d times, expected %d, and didn't call Fatal()",
				t.numErrorCalls,
				numCalls)
		} else if t.numErrorCalls != numCalls {
			t.T.Errorf(
				"the test called Error() %d times, expected %d, but did call Fatal()",
				t.numErrorCalls,
				numCalls)
		} else if !t.fatalCalled {
			t.T.Error("the test did not call Fatal() but Error() the expected amount of times")
		}
	}, fn)
}

func (t *Tester) assert(checkState func(*Tester), fn func(T)) {
	defer checkState(t)

	panicked := true
	defer func() {
		// only recover panic from Fatal()
		if panicked && t.fatalCalled {
			recover()
		}
	}()

	t.numErrorCalls = 0   // reset error state
	t.fatalCalled = false // reset fatal state
	fn(t)                 // call the test with the tester
	panicked = false
}
