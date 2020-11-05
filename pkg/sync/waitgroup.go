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

package sync

import (
	"context"
	"sync"
)

var waitGroupClosedCh chan struct{}

func init() {
	waitGroupClosedCh = make(chan struct{})
	close(waitGroupClosedCh)
}

// WaitGroup is an analog to sync.WaitGroup but also features WaitCh, which
// exposes a channel that is closed once the wait group is fulfilled, and
// WaitCtx, which allows waiting until either the wait group is fulfilled or the
// provided context expires.
type WaitGroup struct {
	mu        sync.Mutex
	remaining int
	done      *Signal
}

// init initialises the wait group, if it was not already.
func (wg *WaitGroup) init() {
	if wg.done == nil {
		wg.done = NewSignal()
	}
}

// Add adds n waiting elements.
func (wg *WaitGroup) Add(n int) {
	wg.mu.Lock()
	defer wg.mu.Unlock()
	wg.init()
	if -n > wg.remaining {
		panic("WaitGroup: negative counter")
	}
	wg.remaining += n
	if wg.remaining == 0 {
		wg.done.Broadcast()
	}
}

// Done decrements the wait counter by one.
func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

// WaitCh returns a channel that will be closed as soon as the wait group is
// fulfilled.
func (wg *WaitGroup) WaitCh() <-chan struct{} {
	wg.mu.Lock()
	defer wg.mu.Unlock()
	wg.init()
	if wg.remaining == 0 {
		return waitGroupClosedCh
	}
	return wg.done.Done()
}

// Wait waits until the wait group is fulfilled.
func (wg *WaitGroup) Wait() {
	<-wg.WaitCh()
}

// WaitCtx waits until the wait group is fulfilled or the context expires.
// Returns whether the wait group was fulfilled.
func (wg *WaitGroup) WaitCtx(ctx context.Context) bool {
	select {
	case <-wg.WaitCh():
		return true
	case <-ctx.Done():
		return false
	}
}
