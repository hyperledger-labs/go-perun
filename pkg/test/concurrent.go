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
	"runtime"
	"strconv"
	"strings"
	stdsync "sync"
	stdatomic "sync/atomic"

	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/pkg/sync/atomic"
)

// ConcT is a testing object used by ConcurrentT stages. It can access the
// parent ConcurrentT's barrier and wait functions. This way, it can wait for
// sibling stages and barriers under the same parent ConcurrentT.
type ConcT struct {
	require.TestingT // The stage's T object.

	ct *ConcurrentT // The parent ConcurrenT object.
}

// Wait waits for a sibling stage or barrier to terminate.
// See ConcurrentT.Wait.
func (t ConcT) Wait(names ...string) {
	t.ct.Wait(names...)
}

// Barrier creates a barrier visible to all sibling stages.
// See ConcurrentT.Barrier.
func (t ConcT) Barrier(name string) {
	t.ct.Barrier(name)
}

// FailBarrier marks a barrier visible to all sibling stages as failed.
// See ConcurrentT.FailBarrier.
func (t ConcT) FailBarrier(name string) {
	defer t.ct.FailBarrier(name)
	t.FailNow()
}

// BarrierN creates a barrier visible to all sibling stages.
// See ConcurrentT.BarrierN.
func (t ConcT) BarrierN(name string, n int) {
	fail := true
	defer func() {
		if fail {
			t.FailNow()
		}
	}()
	t.ct.BarrierN(name, n)
	fail = false
}

// FailBarrierN marks a barrier visible to all sibling stages as failed.
// See ConcurrentT.FailBarrierN.
func (t ConcT) FailBarrierN(name string, n int) {
	defer t.ct.FailBarrierN(name, n)
	t.FailNow()
}

var _ require.TestingT = (*stage)(nil)

// stage is a single stage of execution in a concurrent test.
type stage struct {
	name      string      // The stage's name.
	failed    atomic.Bool // Whether a stage failed.
	spawnOnce stdsync.Once

	wg    sync.WaitGroup // Stage barrier.
	wgN   int            // Used to detect spawn() calls with wrong N.
	count int32          // The number of instances.

	require.TestingT // The stage's test object.

	ct *ConcurrentT // The concurrent testing object.
}

// spawn sets up a stage when it is spawned.
// Checks that the stage is not spawned multiple times.
func (s *stage) spawn(n int) {
	s.spawnOnce.Do(func() {
		s.wg.Add(n - 1)
		s.wgN = n
	})

	if n != s.wgN {
		panic("spawned stage '" + s.name + "' with inconsistent N: " +
			strconv.Itoa(n) + " vs. " + strconv.Itoa(s.wgN))
	}
	if int(stdatomic.AddInt32(&s.count, 1)) > s.wgN {
		panic("spawned stage '" + s.name + "' too often")
	}
}

// pass marks the stage as passed and waits until it is complete.
func (s *stage) pass() {
	s.wg.Done()
}

// FailNow marks the stage as failed and terminates the goroutine.
func (s *stage) FailNow() {
	s.failed.Set()
	defer s.wg.Done()
	s.ct.FailNow()
}

// ConcurrentT is a testing object that can be used in multiple goroutines.
// Specifically, using the helper objects created by the Stage/StageN calls,
// FailNow can be called by any goroutine (however, the helper objects must not
// be used in multiple goroutines).
type ConcurrentT struct {
	failNowMutex sync.Mutex
	t            require.TestingT
	failed       bool
	failedCh     chan struct{}

	mutex  sync.Mutex
	stages map[string]*stage
}

// NewConcurrent creates a new concurrent testing object.
func NewConcurrent(t require.TestingT) *ConcurrentT {
	return &ConcurrentT{
		t:        t,
		stages:   make(map[string]*stage),
		failedCh: make(chan struct{}),
	}
}

// spawnStage retrieves/creates a stage and sets it up.
func (t *ConcurrentT) spawnStage(name string, n int) *stage {
	stage := t.getStage(name)
	stage.spawn(n)
	return stage
}

// getStage retrieves and existing stage or creates a new one, if it does not
// exist yet.
func (t *ConcurrentT) getStage(name string) *stage {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if s, ok := t.stages[name]; ok {
		return s
	}

	s := &stage{name: name, TestingT: t.t, ct: t}
	s.wg.Add(1)
	t.stages[name] = s
	return s
}

// Wait waits until the stages and barriers with the requested names terminate.
// If any stage or barrier fails, terminates the current goroutine or test.
func (t *ConcurrentT) Wait(names ...string) {
	if len(names) == 0 {
		panic("Wait(): called with 0 names")
	}

	for _, name := range names {
		stage := t.getStage(name)
		select {
		case <-stage.wg.WaitCh():
			if stage.failed.IsSet() {
				t.FailNow()
			}
		case <-t.failedCh:
			runtime.Goexit()
		}
	}
}

// FailNow fails and aborts the test.
func (t *ConcurrentT) FailNow() {
	t.failNowMutex.Lock()
	defer t.failNowMutex.Unlock()
	if !t.failed {
		t.failed = true
		defer close(t.failedCh)
		t.t.FailNow()
	} else {
		runtime.Goexit()
	}
}

// StageN creates a named execution stage.
// The parameter goroutines specifies the number of goroutines that share the
// stage. This number must be consistent across all StageN calls with the same
// stage name and exactly match the number of times StageN is called for that
// name.
// Executes fn. If fn calls FailNow on the supplied T object, the stage fails.
// fn must not spawn any goroutines or pass along the T object to goroutines
// that call T.Fatal. To achieve this, make other goroutines call
// ConcurrentT.StageN() instead.
func (t *ConcurrentT) StageN(name string, goroutines int, fn func(ConcT)) {
	stage := t.spawnStage(name, goroutines)

	stageT := ConcT{TestingT: stage, ct: t}
	abort := CheckAbort(func() {
		fn(stageT)
	})

	if abort != nil {
		// Fail the stage, if it had not been marked as such, yet.
		if stage.failed.TrySet() {
			defer stage.wg.Done()
		}
		// If it is a panic or Goexit from certain contexts, print stack trace.
		if _, ok := abort.(*Panic); ok || shouldPrintStack(abort.Stack()) {
			print("\n", abort.String())
		}
		t.FailNow()
	}

	stage.pass()
	t.Wait(name)
}

func shouldPrintStack(stack string) bool {
	// Ignore goroutines that terminate in Wait() because that's a controlled
	// shutdown of the test and not an error.
	return !strings.Contains(stack, "(*ConcurrentT).Wait(")
}

// Stage creates a named execution stage.
// It is a shorthand notation for StageN(name, 1, fn).
func (t *ConcurrentT) Stage(name string, fn func(ConcT)) {
	t.StageN(name, 1, fn)
}

// BarrierN creates a barrier that can be waited on by other goroutines using
// Wait(). After n calls to BarrierN have been made, all waiting goroutines
// continue. Similar to Stage and StageN, all calls to the same barrier must
// share the same n and there must be at most n calls to BarrierN or
// FailBarrierN for each barrier name.
func (t *ConcurrentT) BarrierN(name string, n int) {
	t.spawnStage(name, n).pass()
	t.Wait(name)
}

// FailBarrierN marks a barrier as failed. It terminates the current test and
// all goroutines waiting for the barrier.
func (t *ConcurrentT) FailBarrierN(name string, n int) {
	t.spawnStage(name, n).FailNow()
}

// Barrier is a shorthand notation for Barrier(name, 1).
func (t *ConcurrentT) Barrier(name string) {
	t.spawnStage(name, 1).pass()
}

// FailBarrier creates a synchronisation point and marks it as failed, so that
// waiting goroutines will terminate.
func (t *ConcurrentT) FailBarrier(name string) {
	t.spawnStage(name, 1).FailNow()
}
