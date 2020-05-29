// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

// CheckGoexit tests whether a supplied function calls runtime.Goexit during its
// execution.
// Returns whether the supplied function did call runtime.Goexit.
func CheckGoexit(function func()) bool {
	done := make(chan struct{})
	goexit := true
	go func() {
		defer close(done)
		func() {
			defer func() { recover() }()
			// Call the function to be checked.
			function()
		}()

		// This is executed if there was a panic, but no goexit.
		goexit = false
	}()

	<-done
	return goexit
}
