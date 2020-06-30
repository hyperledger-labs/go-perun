// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// This test tests the test tester, using itself for failing assertions.
func TestTester(_t *testing.T) {
	// positive assertions should not produce an error on _t.
	_t.Run("correct assertions", func(_t *testing.T) {
		tester := NewTester(_t)
		tester.AssertError(func(t T) {
			t.Errorf("")
		})
		tester.AssertError(func(t T) {
			t.Errorf("")
		})
		tester.AssertErrorN(func(t T) {
			t.Errorf("")
			t.Errorf("")
			t.Errorf("")
			t.Errorf("")
			t.Errorf("")
		}, 5)

		tester.AssertFatal(func(t T) {
			t.FailNow()
		})

		tester.AssertErrorFatal(func(t T) {
			t.Errorf("")
			t.FailNow()
		})
		tester.AssertErrorNFatal(func(t T) {
			t.Errorf("")
			t.Errorf("")
			t.Errorf("")
			t.Errorf("")
			t.Errorf("")
			t.FailNow()
		}, 5)
	})

	_t.Run("failing assertions", func(_t *testing.T) {
		// this is the tester with which we test the tester.
		tt := NewTester(_t)

		// not calling Errorf should produce an error.
		tt.AssertError(func(t T) {
			AssertError(t, func(T) {})
		})

		// calling FailNow instead of Errorf should produce an error.
		tt.AssertError(func(t T) {
			AssertError(t, func(t T) { t.FailNow() })
		})

		// calling Errorf 2 times while 3 expected should produce an error.
		tt.AssertError(func(t T) {
			AssertErrorN(t, func(t T) { t.Errorf(""); t.Errorf("") }, 3)
		})

		// not calling FailNow should produce an error.
		tt.AssertError(func(t T) {
			AssertFatal(t, func(T) {})
		})

		// calling Errorf instead of FailNow should produce an error.
		tt.AssertError(func(t T) {
			AssertFatal(t, func(t T) { t.Errorf("") })
		})

		// not calling Errorf or FailNow should produce an error.
		tt.AssertError(func(t T) {
			AssertErrorFatal(t, func(t T) {})
		})

		// calling only Errorf should produce an error.
		tt.AssertError(func(t T) {
			AssertErrorFatal(t, func(t T) { t.Errorf("") })
		})

		// calling only FailNow should produce an error.
		tt.AssertError(func(t T) {
			AssertErrorFatal(t, func(t T) { t.FailNow() })
		})

		// not calling Errorf or FailNow should produce an error.
		tt.AssertError(func(t T) {
			AssertErrorNFatal(t, func(t T) {}, 1)
		})

		// calling only Errorf should produce an error.
		tt.AssertError(func(t T) {
			AssertErrorNFatal(t, func(t T) { t.Errorf(""); t.Errorf("") }, 2)
		})

		// calling only FailNow should produce an error.
		tt.AssertError(func(t T) {
			AssertErrorNFatal(t, func(t T) { t.FailNow() }, 2)
		})

		// calling Errorf the wrong amount of times and FailNow should produce an error.
		tt.AssertError(func(t T) {
			AssertErrorNFatal(t, func(t T) { t.Errorf(""); t.FailNow() }, 2)
		})
	})

	// tests that the tester rethrows panics that are not caused by its fatal().
	_t.Run("panicking assertion", func(_t *testing.T) {
		assert := assert.New(_t)
		tester := NewTester(_t)
		assert.PanicsWithValue(
			"boom",
			func() {
				tester.assert(func(*testerT) {}, func(T) { panic("boom") })
			},
			"Tester.assert caught other panic",
		)

		// panic(nil)
		panicked, pval := CheckPanic(func() { tester.assert(func(*testerT) {}, func(T) { panic(nil) }) })
		assert.True(panicked, "Tester.assert caught panic(nil)")
		assert.Nil(pval, "Tester.assert rethrew panic(nil) as non-nil")
	})
}
