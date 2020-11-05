// Copyright 2020 - See NOTICE file for copyright holders.
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

package sync_test

import (
	"context"
	stdsync "sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/context/test"
	"perun.network/go-perun/pkg/sync"
)

func TestSignal_Signal(t *testing.T) {
	const N = 64

	s := sync.NewSignal()

	var done stdsync.WaitGroup
	waitNTimes(s, &done, N)

	for i := 0; i < N; i++ {
		go s.Signal()
	}

	test.AssertTerminatesQuickly(t, done.Wait)
}

func TestSignal_Broadcast(t *testing.T) {
	const N = 20

	s := sync.NewSignal()

	for x := 0; x < 4; x++ {
		var done stdsync.WaitGroup
		waitNTimes(s, &done, N)

		// ensure that all goroutines are at s.Wait().
		time.Sleep(100 * time.Millisecond)

		s.Broadcast()
		test.AssertTerminatesQuickly(t, done.Wait)
	}
}

func waitNTimes(s *sync.Signal, wg *stdsync.WaitGroup, n int) {
	started := make(chan struct{})
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			started <- struct{}{}
			defer wg.Done()
			s.Wait()
		}()
		// Ensure the coroutine is already started to prevent scheduling delays
		// on slow pipelines.
		<-started
	}
}

func TestSignal_Wait(t *testing.T) {
	s := sync.NewSignal()
	test.AssertNotTerminatesQuickly(t, s.Wait)

	go func() {
		time.Sleep(200 * time.Millisecond)
		s.Broadcast()
	}()
	test.AssertNotTerminates(t, 100*time.Millisecond, s.Wait)
	test.AssertTerminates(t, 200*time.Millisecond, s.Wait)
}

func TestSignal_WaitCtx(t *testing.T) {
	s := sync.NewSignal()

	timeout := 100 * time.Millisecond
	go func() {
		time.Sleep(3 * timeout)
		s.Signal()
	}()
	test.AssertNotTerminates(t, timeout, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*timeout)
		defer cancel()
		assert.False(t, s.WaitCtx(ctx))
	})
	test.AssertTerminates(t, 3*timeout, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*timeout)
		defer cancel()
		assert.True(t, s.WaitCtx(ctx))
	})
}
