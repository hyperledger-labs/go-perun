// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

type (
	// T is part of the interface that testing.T implements. Receive this type
	// instead of *testing.T if you want to test your tests with the Tester.
	T interface {
		Error(...interface{})
		Errorf(string, ...interface{})
		Fatal(...interface{})
		Fatalf(string, ...interface{})

		Helper()
	}

	// Tester is a testing.T mock to test tests.
	// Create new instances of it with NewTester(t), passing it the actual
	// *testing.T that the mock should call if an assertion fails.
	// Then let the tests you want to test receive test.T instead of *testing.T and
	// call your tests inside Tester.AssertX() calls.
	Tester struct {
		// T is the testing object that should be used to report failures of tests.
		// Usually, this is a *testing.T.
		T
	}

	// testerT is the T object that the test tester passes to the to be tested
	// test to record calls to Error and Fatal.
	testerT struct {
		fatalCalled   bool
		numErrorCalls uint
	}
)

// NewTester creates a new test tester, wrapping the passed actual test and
// calling t.Error() on it if a test test fails.
func NewTester(t T) *Tester {
	return &Tester{T: t}
}

func (t *testerT) Error(...interface{}) {
	t.err()
}

func (t *testerT) Errorf(string, ...interface{}) {
	t.err()
}

func (t *testerT) err() {
	t.numErrorCalls++
}

func (t *testerT) Fatal(...interface{}) {
	t.fatal()
}

func (t *testerT) Fatalf(string, ...interface{}) {
	t.fatal()
}

func (t *testerT) fatal() {
	t.fatalCalled = true
	// panic() to stop execution in the test, as would be the case in a test where
	// Fatal is called. The panic is recovered in the Assert... methods.
	panic("testerT.fatal()")
}

func (t *testerT) Helper() {}

// AssertFatal checks that the passed function fn calls T.Fatal() on the T
// object it calls fn with.
func (t *Tester) AssertFatal(fn func(T)) {
	t.assert(func(tt *testerT) {
		if !tt.fatalCalled {
			t.T.Error("the test did not call Fatal()")
		}
	}, fn)
}

// AssertError checks that the passed function fn calls T.Error() on the T
// object it calls fn with.
func (t *Tester) AssertError(fn func(T)) {
	t.assert(func(tt *testerT) {
		if tt.numErrorCalls == 0 {
			t.T.Error("the test did not call Error()")
		}
	}, fn)
}

// AssertErrorN checks that the passed function fn calls T.Error() numCalls
// times on the T object it calls fn with.
func (t *Tester) AssertErrorN(fn func(T), numCalls uint) {
	t.assert(func(tt *testerT) {
		if tt.numErrorCalls != numCalls {
			t.T.Errorf("the test called Error() %d times, expected %d", tt.numErrorCalls, numCalls)
		}
	}, fn)
}

// AssertErrorFatal checks that the passed function fn calls T.Error() and
// T.Fatal() on the T object it calls fn with.
func (t *Tester) AssertErrorFatal(fn func(T)) {
	t.assert(func(tt *testerT) {
		if tt.numErrorCalls == 0 && !tt.fatalCalled {
			t.T.Error("the test did neither call Error() nor Fatal()")
		} else if tt.numErrorCalls == 0 {
			t.T.Error("the test did not call Error() but Fatal()")
		} else if !tt.fatalCalled {
			t.T.Error("the test did not call Fatal() but Error()")
		}
	}, fn)
}

// AssertErrorNFatal checks that the passed function fn calls T.Error() numCalls
// times and T.Fatal() on the T object it calls fn with.
func (t *Tester) AssertErrorNFatal(fn func(T), numCalls uint) {
	t.assert(func(tt *testerT) {
		if tt.numErrorCalls != numCalls && !tt.fatalCalled {
			t.T.Errorf(
				"the test called Error() %d times, expected %d, and didn't call Fatal()",
				tt.numErrorCalls,
				numCalls)
		} else if tt.numErrorCalls != numCalls {
			t.T.Errorf(
				"the test called Error() %d times, expected %d, but did call Fatal()",
				tt.numErrorCalls,
				numCalls)
		} else if !tt.fatalCalled {
			t.T.Error("the test did not call Fatal() but Error() the expected amount of times")
		}
	}, fn)
}

func (t *Tester) assert(check func(*testerT), fn func(T)) {
	tt := new(testerT)
	defer check(tt)

	panicked := true
	defer func() {
		// only recover panic from Fatal()
		if panicked && tt.fatalCalled {
			recover()
		}
	}()

	fn(tt) // call the test with the testerT to record calls
	panicked = false
}

// AssertFatal checks that the passed function fn calls T.Fatal() on the T
// object it calls fn with. Errors are reported on t.
func AssertFatal(t T, fn func(T)) {
	NewTester(t).AssertFatal(fn)
}

// AssertError checks that the passed function fn calls T.Error() on the T
// object it calls fn with. Errors are reported on t.
func AssertError(t T, fn func(T)) {
	NewTester(t).AssertError(fn)
}

// AssertErrorN checks that the passed function fn calls T.Error() numCalls
// times on the T object it calls fn with. Errors are reported on t.
func AssertErrorN(t T, fn func(T), numCalls uint) {
	NewTester(t).AssertErrorN(fn, numCalls)
}

// AssertErrorFatal checks that the passed function fn calls T.Error() and
// T.Fatal() on the T object it calls fn with. Errors are reported on t.
func AssertErrorFatal(t T, fn func(T)) {
	NewTester(t).AssertErrorFatal(fn)
}

// AssertErrorNFatal checks that the passed function fn calls T.Error() numCalls
// times and T.Fatal() on the T object it calls fn with. Errors are reported on t.
func AssertErrorNFatal(t T, fn func(T), numCalls uint) {
	NewTester(t).AssertErrorNFatal(fn, numCalls)
}
