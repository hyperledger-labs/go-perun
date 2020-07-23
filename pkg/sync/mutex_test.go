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

package sync

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const timeout = 200 * time.Millisecond

// TestMutex_Lock tests that an empty mutex can be locked.
func TestMutex_Lock(t *testing.T) {
	t.Parallel()

	var m Mutex

	done := make(chan struct{}, 1)
	go func() {
		m.Lock()
		done <- struct{}{}
	}()

	select {
	case <-done:
		assert.NotPanics(t, func() { m.Unlock() }, "Unlock must succeed")
	case <-time.NewTimer(timeout).C:
		t.Error("lock on new mutex did not instantly succeed")
	}
}

// TestMutex_TryLock tests that TryLock() can lock an empty mutex, and that
// locked mutexes cannot be locked again.
func TestMutex_TryLock(t *testing.T) {
	t.Parallel()

	var m Mutex
	// Try instant lock without context.
	assert.True(t, m.TryLock(), "TryLock() on new mutex must succeed")
	assert.False(t, m.TryLock(), "TryLock() on locked mutex must fail")
	assert.NotPanics(t, func() { m.Unlock() }, "Unlock must succeed")
}

// TestMutex_TryLockCtx_Nil tests that TryLockCtx(nil) behaves like TryLock().
func TestMutex_TryLockCtx_Nil(t *testing.T) {
	t.Parallel()

	var m Mutex
	// Try instant lock without context.
	assert.True(t, m.TryLockCtx(nil), "TryLock() on new mutex must succeed")
	assert.False(t, m.TryLockCtx(nil), "TryLock() on locked mutex must fail")
	assert.NotPanics(t, func() { m.Unlock() }, "Unlock must succeed")
}

// TestMutex_TryLockCtx_DoneContext tests that a cancelled context can never be
// used to acquire the mutex.
func TestMutex_TryLockCtx_DoneContext(t *testing.T) {
	t.Parallel()

	var m Mutex
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Try often because of random `select` case choices.
	for i := 0; i < 256; i++ {
		assert.False(t, m.TryLockCtx(ctx), "TryLockCtx on closed context must fail")
	}
}

// TestMutex_TryLockCtx_WithTimeout tests that the context's timeout is
// properly adhered to.
func TestMutex_TryLockCtx_WithTimeout(t *testing.T) {
	t.Parallel()
	testMutexTryLockCtx(t, 1, 2, true)
}

// TestMutex_TryLockCtx_WithTimeout_Fail tests that TryLockCtx fails if it
// times out.
func TestMutex_TryLockCtx_WithTimeout_Fail(t *testing.T) {
	t.Parallel()
	testMutexTryLockCtx(t, 2, 1, false)
}

func testMutexTryLockCtx(
	t *testing.T,
	unlockDelay time.Duration,
	lockTimeout time.Duration,
	expectSuccess bool) {

	t.Helper()
	unlockDelay *= timeout
	lockTimeout *= timeout

	var m Mutex
	m.Lock()

	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()
	done := make(chan bool, 1)
	go func() { done <- m.TryLockCtx(ctx) }()
	go func() {
		<-time.NewTimer(unlockDelay).C
		m.Unlock()
	}()

	// Check that it does not return early.
	select {
	case <-time.NewTimer(unlockDelay / 2).C:
	case success := <-done:
		assert.False(t, expectSuccess,
			"TryLockCtx should not have returned before unlocking")
		assert.False(t, success, "TryLockCtx should have failed")
	}

	// Check result.
	select {
	case <-ctx.Done():
		assert.False(t, expectSuccess, "TryLockCtx was not supposed to time out")
	case success := <-done:
		if expectSuccess {
			assert.True(t, success, "TryLockCtx should have succeeded")
		} else {
			assert.False(t, success, "TryLockCtx should have failed")
		}
	}
}

// TestMutex_Unlock tests that unlocking a locked mutex will make it lockable
// again.
func TestMutex_Unlock(t *testing.T) {
	t.Parallel()

	var m Mutex
	assert.Panics(t, func() { m.Unlock() }, "uninitialized Unlock must panic")
	m.Lock()
	assert.NotPanics(t, func() { m.Unlock() }, "Unlock must succeed")
	assert.Panics(t, func() { m.Unlock() }, "double Unlock must panic")
	assert.True(t, m.TryLock(), "Unlock must make the next TryLock succeed")
}
