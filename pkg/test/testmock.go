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

import "runtime"

type (
	// T is part of the interface that testing.T implements. Receive this type
	// instead of *testing.T if you want to test your tests with the Tester.
	T interface {
		Errorf(string, ...interface{})
		FailNow()

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
	// test to record calls to Errorf and FailNow.
	testerT struct {
		failNowCalled bool
		numErrorCalls uint
	}
)

// NewTester creates a new test tester, wrapping the passed actual test and
// calling t.Errorf() on it if a test test fails.
func NewTester(t T) *Tester {
	return &Tester{T: t}
}

func (t *testerT) Errorf(string, ...interface{}) {
	t.numErrorCalls++
}

func (t *testerT) FailNow() {
	t.failNowCalled = true
	// runtime.Goexit() to stop execution in the test, as would be the case in
	// a test where Fatal is called. The Goexit is stopped in the Assert...
	// methods.
	runtime.Goexit()
}

func (t *testerT) Helper() {}

// AssertFatal checks that the passed function fn calls T.FailNow() on the T
// object it calls fn with.
func (t *Tester) AssertFatal(fn func(T)) {
	t.assert(func(tt *testerT) {
		if !tt.failNowCalled {
			t.T.Errorf("the test did not call FailNow()")
		}
	}, fn)
}

// AssertError checks that the passed function fn calls T.Errorf() on the T
// object it calls fn with.
func (t *Tester) AssertError(fn func(T)) {
	t.assert(func(tt *testerT) {
		if tt.numErrorCalls == 0 {
			t.T.Errorf("the test did not call Errorf()")
		}
	}, fn)
}

// AssertErrorN checks that the passed function fn calls T.Errorf() numCalls
// times on the T object it calls fn with.
func (t *Tester) AssertErrorN(fn func(T), numCalls uint) {
	t.assert(func(tt *testerT) {
		if tt.numErrorCalls != numCalls {
			t.T.Errorf("the test called Errorf() %d times, expected %d", tt.numErrorCalls, numCalls)
		}
	}, fn)
}

// AssertErrorFatal checks that the passed function fn calls T.Errorf() and
// T.FailNow() on the T object it calls fn with.
func (t *Tester) AssertErrorFatal(fn func(T)) {
	t.assert(func(tt *testerT) {
		// nolint: gocritic
		if tt.numErrorCalls == 0 && !tt.failNowCalled {
			t.T.Errorf("the test did neither call Errorf() nor FailNow()")
		} else if tt.numErrorCalls == 0 {
			t.T.Errorf("the test did not call Errorf() but FailNow()")
		} else if !tt.failNowCalled {
			t.T.Errorf("the test did not call FailNow() but Errorf()")
		}
	}, fn)
}

// AssertErrorNFatal checks that the passed function fn calls T.Errorf() numCalls
// times and T.FailNow() on the T object it calls fn with.
func (t *Tester) AssertErrorNFatal(fn func(T), numCalls uint) {
	t.assert(func(tt *testerT) {
		// nolint: gocritic
		if tt.numErrorCalls != numCalls && !tt.failNowCalled {
			t.T.Errorf(
				"the test called Errorf() %d times, expected %d, and didn't call FailNow()",
				tt.numErrorCalls,
				numCalls)
		} else if tt.numErrorCalls != numCalls {
			t.T.Errorf(
				"the test called Errorf() %d times, expected %d, but did call FailNow()",
				tt.numErrorCalls,
				numCalls)
		} else if !tt.failNowCalled {
			t.T.Errorf("the test did not call FailNow() but Errorf() the expected amount of times")
		}
	}, fn)
}

func (t *Tester) assert(check func(*testerT), fn func(T)) {
	tt := new(testerT)
	defer check(tt)

	if CheckGoexit(func() { fn(tt) }) && !tt.failNowCalled {
		runtime.Goexit()
	}
}

// AssertFatal checks that the passed function fn calls T.FailNow() on the T
// object it calls fn with. Errors are reported on t.
func AssertFatal(t T, fn func(T)) {
	NewTester(t).AssertFatal(fn)
}

// AssertError checks that the passed function fn calls T.Errorf() on the T
// object it calls fn with. Errors are reported on t.
func AssertError(t T, fn func(T)) {
	NewTester(t).AssertError(fn)
}

// AssertErrorN checks that the passed function fn calls T.Errorf() numCalls
// times on the T object it calls fn with. Errors are reported on t.
func AssertErrorN(t T, fn func(T), numCalls uint) {
	NewTester(t).AssertErrorN(fn, numCalls)
}

// AssertErrorFatal checks that the passed function fn calls T.Errorf() and
// T.FailNow() on the T object it calls fn with. Errors are reported on t.
func AssertErrorFatal(t T, fn func(T)) {
	NewTester(t).AssertErrorFatal(fn)
}

// AssertErrorNFatal checks that the passed function fn calls T.Errorf() numCalls
// times and T.FailNow() on the T object it calls fn with. Errors are reported on t.
func AssertErrorNFatal(t T, fn func(T), numCalls uint) {
	NewTester(t).AssertErrorNFatal(fn, numCalls)
}
