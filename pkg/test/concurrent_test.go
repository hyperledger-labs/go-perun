// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStage_FailNow(t *testing.T) {
	t.Run("first fail", func(t *testing.T) {
		AssertFatal(t, func(t T) {
			ct := NewConcurrent(newFailNowTester(t))
			s := ct.spawnStage("stage", 1)
			s.FailNow()
		})
	})

	t.Run("second fail", func(t *testing.T) {
		ct := NewConcurrent(nil)
		ct.failed = true
		s := ct.spawnStage("stage", 1)
		assert.True(t, CheckGoexit(s.FailNow))
	})
}

func TestConcurrentT_Wait(t *testing.T) {
	t.Run("0 names", func(t *testing.T) {
		ct := NewConcurrent(t)
		require.Panics(t, func() { ct.Wait() })
	})

	t.Run("unknown name", func(t *testing.T) {
		ct := NewConcurrent(t)
		ct.Stage("known", func(t require.TestingT) {
		})
		AssertNotTerminates(t, timeout, func() { ct.Wait("unknown") })
	})

	t.Run("known name", func(t *testing.T) {
		ct := NewConcurrent(t)
		go ct.Stage("known", func(t require.TestingT) {
			time.Sleep(timeout / 2)
		})
		AssertTerminates(t, timeout, func() { ct.Wait("known") })
	})

	t.Run("failed stage", func(t *testing.T) {
		ct := NewConcurrent(t)
		s := ct.spawnStage("stage", 1)
		s.failed.Set()
		s.pass()

		assert.True(t, CheckGoexit(func() { ct.Wait("stage") }),
			"Waiting for a failed stage must call runtime.Goexit.")
	})
}

type failNowTester struct {
	*testerT
}

func newFailNowTester(t T) *failNowTester {
	return &failNowTester{testerT: t.(*testerT)}
}

func (t *failNowTester) FailNow() { t.fatal() }

func TestConcurrentT_FailNow(t *testing.T) {
	var ct *ConcurrentT

	// Test that NewConcurrent.FailNow() calls T.FailNow().
	AssertFatal(t, func(t T) {
		ct = NewConcurrent(newFailNowTester(t))
		ct.FailNow()
	})

	// Test that after that, FailNow() calls runtime.Goexit().
	assert.True(t, CheckGoexit(ct.FailNow),
		"redundant FailNow() must call runtime.Goexit()")
}

func TestConcurrentT_StageN(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		ct := NewConcurrent(t)
		var executed, returned sync.WaitGroup
		executed.Add(2)
		returned.Add(2)

		for i := 0; i < 2; i++ {
			go func() {
				ct.StageN("stage", 2, func(t require.TestingT) {
					executed.Done()
				})
				returned.Done()
			}()
		}

		AssertTerminates(t, timeout, executed.Wait)
		AssertTerminates(t, timeout, returned.Wait)
	})

	t.Run("n*m happy", func(t *testing.T) {
		N := 100
		M := 100

		ct := NewConcurrent(t)

		for g := 0; g < N; g++ {
			go func(g int) {
				for stage := 0; stage < M; stage++ {
					if g&1 == 0 {
						ct.StageN(strconv.Itoa(stage), N/2, func(t require.TestingT) {
						})
					} else {
						ct.Wait(strconv.Itoa(stage))
					}
				}
			}(g)
		}
	})

	t.Run("n*m sad", func(t *testing.T) {
		N := 100
		M := 100
		AssertFatal(t, func(t T) {
			ct := NewConcurrent(newFailNowTester(t))
			var wg sync.WaitGroup
			wg.Add(N)
			for g := 0; g < N; g++ {
				go func(g int) {
					defer wg.Done()
					for stage := 0; stage < M; stage++ {
						ct.StageN(strconv.Itoa(stage), N, func(t require.TestingT) {
							if g == N/2 {
								t.FailNow()
							}
						})
					}
				}(g)
			}

			wg.Wait()
		})
	})

	t.Run("too few goroutines", func(t *testing.T) {
		ct := NewConcurrent(t)
		AssertNotTerminates(t, timeout, func() {
			ct.StageN("stage", 2, func(t require.TestingT) {})
		})
	})

	t.Run("too many goroutines", func(t *testing.T) {
		ct := NewConcurrent(t)
		go ct.StageN("stage", 2, func(t require.TestingT) {})
		ct.StageN("stage", 2, func(t require.TestingT) {})
		assert.Panics(t, func() {
			ct.StageN("stage", 2, func(t require.TestingT) {})
		})
	})

	t.Run("inconsistent N", func(t *testing.T) {
		ct := NewConcurrent(t)
		var created sync.WaitGroup
		created.Add(1)

		go ct.StageN("stage", 2, func(t require.TestingT) {
			created.Done()
		})

		created.Wait()
		assert.Panics(t, func() {
			ct.StageN("stage", 3, func(t require.TestingT) {})
		})
	})

	t.Run("panic", func(t *testing.T) {
		AssertFatal(t, func(t T) {
			ct := NewConcurrent(newFailNowTester(t))
			ct.Stage("stage", func(require.TestingT) { panic(nil) })
		})
	})
}
