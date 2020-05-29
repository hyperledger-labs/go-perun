// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

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
