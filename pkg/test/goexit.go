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
