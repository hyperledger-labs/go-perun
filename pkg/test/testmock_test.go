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
	// positive assertions should not produce an error on _t
	_t.Run("correct assertions", func(_t *testing.T) {
		tester := NewTester(_t)
		tester.AssertError(func(t T) {
			t.Error()
		})
		tester.AssertError(func(t T) {
			t.Errorf("")
		})
		tester.AssertErrorN(func(t T) {
			t.Error()
			t.Errorf("")
			t.Error()
			t.Errorf("")
			t.Error()
		}, 5)

		tester.AssertFatal(func(t T) {
			t.Fatal()
		})
		tester.AssertFatal(func(t T) {
			t.Fatalf("")
		})

		tester.AssertErrorFatal(func(t T) {
			t.Errorf("")
			t.Error()
			t.Fatal()
		})
		tester.AssertErrorNFatal(func(t T) {
			t.Error()
			t.Errorf("")
			t.Error()
			t.Errorf("")
			t.Error()
			t.Fatal()
		}, 5)
	})

	_t.Run("failing assertions", func(_t *testing.T) {
		// this is the tester with which we test the tester.
		tt := NewTester(_t)

		// not calling Error should produce an error
		tt.AssertError(func(t T) {
			AssertError(t, func(T) {})
		})

		// calling Fatal instead of Error should produce an error
		tt.AssertError(func(t T) {
			AssertError(t, func(t T) { t.Fatal() })
		})

		// calling Error 2 times while 3 expected should produce an error
		tt.AssertError(func(t T) {
			AssertErrorN(t, func(t T) { t.Error(); t.Errorf("") }, 3)
		})

		// not calling Fatal should produce an error
		tt.AssertError(func(t T) {
			AssertFatal(t, func(T) {})
		})

		// calling Error instead of Fatal should produce an error
		tt.AssertError(func(t T) {
			AssertFatal(t, func(t T) { t.Error() })
		})

		// not calling Error or Fatal should produce an error
		tt.AssertError(func(t T) {
			AssertErrorFatal(t, func(t T) {})
		})

		// calling only Error should produce an error
		tt.AssertError(func(t T) {
			AssertErrorFatal(t, func(t T) { t.Error() })
		})

		// calling only Fatal should produce an error
		tt.AssertError(func(t T) {
			AssertErrorFatal(t, func(t T) { t.Fatal() })
		})

		// not calling Error or Fatal should produce an error
		tt.AssertError(func(t T) {
			AssertErrorNFatal(t, func(t T) {}, 1)
		})

		// calling only Error should produce an error
		tt.AssertError(func(t T) {
			AssertErrorNFatal(t, func(t T) { t.Error(); t.Error() }, 2)
		})

		// calling only Fatal should produce an error
		tt.AssertError(func(t T) {
			AssertErrorNFatal(t, func(t T) { t.Fatal() }, 2)
		})

		// calling Error the wrong amount of times and Fatal should produce an error
		tt.AssertError(func(t T) {
			AssertErrorNFatal(t, func(t T) { t.Error(); t.Fatal() }, 2)
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
