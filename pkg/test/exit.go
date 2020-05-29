// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

type (
	// Exit can test calls to an exit function, which is usually a package global
	// set to os.Exit by default.
	Exit struct {
		t    T
		exit *func(int)
	}

	// An exiter is used by Exit's Assert methods to record calls to exit.
	exiter struct {
		called bool
		code   int
	}
)

// NewExit creates a new Exit tester. The exit function pointer should usually
// be the address of a global exit function variable that is set to os.Exit by
// default. The exit function is temporarily modified during Assert tests.
func NewExit(t T, exit *func(int)) *Exit {
	return &Exit{
		t:    t,
		exit: exit,
	}
}

// Exit records exit calls.
func (e *exiter) Exit(code int) {
	e.called = true
	e.code = code
	// panic() to stop the execution of the function, as would a normal call to
	// os.Exit() do. Unfortunately, we cannot prevent stack unwinding but this is
	// as close as we can get.
	panic("exiter.Exit()")
}

// AssertExit asserts that fn calls the exit function that was passed to NewExit
// with the given code. Usually, exit is a package global function variable that
// is set to os.Exit by default. The exit function is temporarily modified
// during the test.
func (e *Exit) AssertExit(fn func(), code int) {
	e.assert(func(ex *exiter) {
		if !ex.called {
			e.t.Error("exit was not called")
		} else if ex.code != code {
			e.t.Errorf("exit was called with wrong code %d, expected %d", ex.code, code)
		}
	}, fn)
}

// AssertNoExit asserts that fn does not call the exit function that was passed
// to NewExit. Usually, exit is a package global function variable that is set
// to os.Exit by default. The exit function is temporarily modified during the
// test.
func (e *Exit) AssertNoExit(fn func()) {
	e.assert(func(ex *exiter) {
		if ex.called {
			e.t.Error("exit was called")
		}
	}, fn)
}

// assert executes function fn and then calls the check function on the *Exit
// tester. The check should check the desired state of the exiter after
// execution of fn, possibly calling e.t.Error() on errors. Panics caused by
// the exiter are recovered.
func (e *Exit) assert(check func(*exiter), fn func()) {
	ex := new(exiter)
	exitBackup := *e.exit
	defer func() { *e.exit = exitBackup }()
	*e.exit = ex.Exit // temporarily record exit calls using the exiter

	defer check(ex)

	panicked := true
	defer func() {
		// check that this panic came from the exiter and let it bubble up the stack o/w
		if panicked && ex.called {
			recover()
		}
	}()

	fn()
	panicked = false
}

// AssertExit asserts that fn calls the provided exit function with the given
// code. Usually, exit is a package global function variable that is set to
// os.Exit by default. The exit function is temporarily modified during the
// test.
func AssertExit(t T, exit *func(int), fn func(), code int) {
	NewExit(t, exit).AssertExit(fn, code)
}

// AssertNoExit asserts that fn does not call the provided exit function.
// Usually, exit is a package global function variable that is set to os.Exit by
// default. The exit function is temporarily modified during the test.
func AssertNoExit(t T, exit *func(int), fn func()) {
	NewExit(t, exit).AssertNoExit(fn)
}
