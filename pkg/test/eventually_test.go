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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventually(t *testing.T) {
	// default parameters for this test
	within, pause := 40*time.Millisecond, 10*time.Millisecond
	numTries := (int)(within/pause) + 1
	et := NewEventually(within, pause) // bind parameters

	t.Run("successful", func(t *testing.T) {
		t.Parallel()

		et.Eventually(t, func(t T) {})

		fails := []func(T){
			func(t T) { t.Errorf("") },
			func(t T) { t.FailNow() },
		}

		for _, fn := range fails {
			ftest := failUntil(time.Now().Add(within), fn)
			et.Eventually(t, ftest.Fail)
			assert.GreaterOrEqual(t, numTries, ftest.NumTries)
			assert.Equal(t, 1, ftest.NumNoPanic,
				"all test calls should panic after first error call")
		}
	})

	t.Run("failing", func(t *testing.T) {
		t.Parallel()
		tt := NewTester(t)

		tt.AssertErrorN(func(t T) {
			et.Eventually(t, func(t T) { t.Errorf("") })
		}, 1)
		tt.AssertErrorN(func(t T) {
			et.Eventually(t, func(t T) { t.Errorf("") })
		}, 1)

		tt.AssertFatal(func(t T) {
			et.Eventually(t, func(t T) { t.FailNow() })
		})

		// fail just until after `within`
		ftest := failUntil(time.Now().Add(within+pause), func(t T) { t.Errorf("") })
		tt.AssertErrorN(func(t T) {
			et.Eventually(t, ftest.Fail)
		}, 1)
		assert.GreaterOrEqual(t, numTries, ftest.NumTries)
		assert.Equal(t, 1, ftest.NumNoPanic,
			"all test calls should panic after first error call")
	})

	// edge cases of time parameters
	t.Run("edge cases", func(t *testing.T) {
		t.Parallel()
		tt := NewTester(t)

		tests := []struct {
			within   time.Duration
			pause    time.Duration
			numCalls uint
		}{
			{0, 0, 1},
			{0, time.Millisecond, 1},
			{time.Millisecond, time.Millisecond, 2},
			{time.Millisecond, 2 * time.Millisecond, 2},
		}

		for _, etest := range tests {
			numCalls := uint(0)
			tt.AssertErrorN(func(t T) {
				Eventually(t, func(t T) {
					numCalls++
					t.Errorf("")
				}, etest.within, etest.pause)
			}, 1)
			assert.Equal(t, etest.numCalls, numCalls)
		}
	})
}

type failer struct {
	ts         time.Time
	failFn     func(T)
	NumTries   int
	NumNoPanic int
}

func (f *failer) Fail(t T) {
	f.NumTries++
	if time.Now().Before(f.ts) {
		f.failFn(t)
	}
	f.NumNoPanic++
}

func failUntil(ts time.Time, failFn func(T)) *failer {
	return &failer{
		ts:     ts,
		failFn: failFn,
	}
}
