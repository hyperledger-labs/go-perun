// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package test contains helper functions for testing.
package test // import "perun.network/go-perun/pkg/test"

// CheckPanic tests whether a supplied function panics during its execution.
// Returns nil if the supplied function did not panic, otherwise, returns the
// value passed to panic().
func CheckPanic(function func()) (value interface{}) {
	defer func() { value = recover() }()
	function()
	return
}
