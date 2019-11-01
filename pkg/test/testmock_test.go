// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

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
		// this is the tester whith which we test the tester.
		tt := NewTester(_t)
		// we create a single tester and call all tests on it in a row.
		tester := NewTester(tt)

		// not calling Error should produce an error
		tt.AssertError(func(T) {
			tester.AssertError(func(T) {})
		})

		// calling Fatal instead of Error should produce an error
		tt.AssertError(func(T) {
			tester.AssertError(func(t T) { t.Fatal() })
		})

		// calling Error 2 times while 3 expected should produce an error
		tt.AssertError(func(T) {
			tester.AssertErrorN(func(t T) { t.Error(); t.Errorf("") }, 3)
		})

		// not calling Fatal should produce an error
		tt.AssertError(func(T) {
			tester.AssertFatal(func(T) {})
		})

		// calling Error instead of Fatal should produce an error
		tt.AssertError(func(T) {
			tester.AssertFatal(func(t T) { t.Error() })
		})

		// not calling Error or Fatal should produce an error
		tt.AssertError(func(T) {
			tester.AssertErrorFatal(func(t T) {})
		})

		// calling only Error should produce an error
		tt.AssertError(func(T) {
			tester.AssertErrorFatal(func(t T) { t.Error() })
		})

		// calling only Fatal should produce an error
		tt.AssertError(func(T) {
			tester.AssertErrorFatal(func(t T) { t.Fatal() })
		})

		// not calling Error or Fatal should produce an error
		tt.AssertError(func(T) {
			tester.AssertErrorNFatal(func(t T) {}, 1)
		})

		// calling only Error should produce an error
		tt.AssertError(func(T) {
			tester.AssertErrorNFatal(func(t T) { t.Error(); t.Error() }, 2)
		})

		// calling only Fatal should produce an error
		tt.AssertError(func(T) {
			tester.AssertErrorNFatal(func(t T) { t.Fatal() }, 2)
		})

		// calling Error the wrong amount of times and Fatal should produce an error
		tt.AssertError(func(T) {
			tester.AssertErrorNFatal(func(t T) { t.Error(); t.Fatal() }, 2)
		})
	})

	// tests that the tester rethrows panics that are not caused by its fatal().
	_t.Run("panicking assertion", func(_t *testing.T) {
		assert := assert.New(_t)
		tester := NewTester(_t)
		assert.PanicsWithValue(
			"boom",
			func() {
				tester.assert(func(*Tester) {}, func(T) { panic("boom") })
			},
			"Tester.assert caught other panic",
		)

		// panic(nil)
		panicked, pval := CheckPanic(func() { tester.assert(func(*Tester) {}, func(T) { panic(nil) }) })
		assert.True(panicked, "Tester.assert caught panic(nil)")
		assert.Nil(pval, "Tester.assert rethrew panic(nil) as non-nil")
	})
}
