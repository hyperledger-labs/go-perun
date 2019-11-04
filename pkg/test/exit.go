// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

// Exit can test calls to an exit function, which is usually a package global
// set to os.Exit by default.
type Exit struct {
	t      T
	called bool
	code   int
}

// NewExit creates a new Exit tester, setting the passed function exit to its
// Exit function. This exit should usually be the address of a global exit
// function that is set to os.Exit by default.
func NewExit(t T, exit *func(int)) *Exit {
	et := &Exit{t: t}
	*exit = et.Exit
	return et
}

// Exit mocks exit calls. Usually, this shouldn't need to be called directly as
// the exit function to patch is set during NewExit().
func (e *Exit) Exit(code int) {
	e.called = true
	e.code = code
	// panic() to stop the execution of the function, as would a normal call to
	// os.Exit() do. Unfortunately, we cannot prevent stack unwinding but this is
	// as close as we can get.
	panic("Exit.Exit()")
}

// AssertExit asserts that fn calls e.Exit(). Usually, fn should call a global
// exit function variable that is set to os.Exit by default.
func (e *Exit) AssertExit(fn func(), code int) {
	e.assert(func(e *Exit) {
		if !e.called {
			e.t.Error("exit was not called")
		} else if e.code != code {
			e.t.Errorf("exit was called with wrong code %d, expected %d", e.code, code)
		}
	}, fn)
}

// AssertNoExit asserts that fn calls e.Exit(). Usually, fn should call a global
// exit function variable that is set to os.Exit by default.
func (e *Exit) AssertNoExit(fn func()) {
	e.assert(func(e *Exit) {
		if e.called {
			e.t.Error("exit was called")
		}
	}, fn)
}

// assert executes function fn and then calls the check function on the *Exit
// tester. The check should check the desired state of *Exit after execution of
// fn, possibly calling t.T.Error() on errors.
// Panics caused by Exit() are recovered.
func (e *Exit) assert(check func(*Exit), fn func()) {
	defer check(e)

	panicked := true
	defer func() {
		// check that this panic came from Exit() and let it bubble up the stack o/w
		if panicked && e.called {
			recover()
		}
	}()

	e.called = false // reset called state
	fn()
	panicked = false
}
