// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExit(t *testing.T) {
	t.Run("positive tests", func(t *testing.T) {
		var exit func(int) // use local instead of global exit for this test
		et := NewExit(t, &exit)
		et.AssertExit(func() { exit(42) }, 42)
		et.AssertNoExit(func() {})

		called := false
		et.AssertExit(func() { exit(42); called = true }, 42)
		assert.False(t, called, "Exit.Exit should pretend further execution")
		assert.PanicsWithValue(
			t,
			"boom",
			func() {
				et.assert(func(*Exit) {}, func() { panic("boom") })
			},
			"Exit.assert() should rethrow panic not caused by Exit()",
		)
	})

	t.Run("failing tests", func(t *testing.T) {
		tt := NewTester(t)
		var exit func(int) // use local instead of global exit for this test
		et := NewExit(tt, &exit)
		// calling AssertExit without exit call should fail
		tt.AssertError(func(T) { et.AssertExit(func() {}, 42) })
		// calling AssertExit with wrong exit code call should fail
		tt.AssertError(func(T) { et.AssertExit(func() { exit(53) }, 42) })
		// calling AssertNoExit with exit call should fail
		tt.AssertError(func(T) { et.AssertNoExit(func() { exit(53) }) })
	})

}
