// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package test contains helper functions for testing.
package test // import "perun.network/go-perun/pkg/test"

// CheckPanic tests whether a supplied function panics during its execution.
// Returns whether the supplied function did panic, and if so, also returns the
// value passed to panic().
func CheckPanic(function func()) (didPanic bool, value interface{}) {
	// Catch the panic, if it happens and store the passed value.
	defer func() { value = recover() }()
	// Set up for the panic case.
	didPanic = true
	// Call the function to be checked.
	function()
	// This is only executed if panic() was not called.
	didPanic = false
	return
}
