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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExit(t *testing.T) {
	t.Run("positive tests", func(t *testing.T) {
		var code int
		exit := func(i int) { code = i } // use local instead of global exit for this test
		et := NewExit(t, &exit)
		et.AssertExit(func() { exit(42) }, 42)
		et.AssertNoExit(func() {})

		code = 0
		exit(17)
		assert.Equal(t, 17, code,
			"Exit tester should not permanently modify the exit function")

		called := false
		et.AssertExit(func() { exit(42); called = true }, 42)
		assert.False(t, called, "calling exit should prevent further execution")

		assert.PanicsWithValue(
			t,
			"boom",
			func() {
				et.assert(func(*exiter) {}, func() { panic("boom") })
			},
			"Exit.assert() should rethrow panic not caused by Exit()",
		)
	})

	t.Run("failing tests", func(t *testing.T) {
		tt := NewTester(t)
		var exit func(int) // use local instead of global exit for this test
		// calling AssertExit without exit call should fail
		tt.AssertError(func(t T) { AssertExit(t, &exit, func() {}, 42) })
		// calling AssertExit with wrong exit code call should fail
		tt.AssertError(func(t T) { AssertExit(t, &exit, func() { exit(53) }, 42) })
		// calling AssertNoExit with exit call should fail
		tt.AssertError(func(t T) { AssertNoExit(t, &exit, func() { exit(53) }) })
	})

}
