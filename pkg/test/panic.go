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

// Package test contains helper functions for testing.
package test

// CheckPanic tests whether a supplied function panics during its execution.
// Returns whether the supplied function did panic, and if so, also returns the
// value passed to panic().
func CheckPanic(function func()) (didPanic bool, value interface{}) {
	// Catch the panic, if it happens and store the passed value.
	defer func() {
		value = recover()
		if p, ok := value.(*Panic); ok {
			value = p.Value()
		}
	}()
	// Set up for the panic case.
	didPanic = true
	// Call the function to be checked.
	function()
	// This is only executed if panic() was not called.
	didPanic = false
	return
}
