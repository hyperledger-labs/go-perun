// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"os"

	"bou.ke/monkey"
)

// CheckExit checks whether a supplied function calls os.Exit() and returns
// whether it exited, and if so, which status code it exited with. Internally
// replaces os.Exit() with a call to panic(), so the supplied function should
// not recover. Due to the use of monkey patching, this function should only be
// used for testing purposes!
func CheckExit(function func()) (exits bool, code int) {
	defer monkey.Unpatch(os.Exit)
	monkey.Patch(os.Exit, func(status int) {
		exits = true
		code = status
		panic(nil)
	})

	// Recover the potential panic.
	CheckPanic(function)
	return
}
