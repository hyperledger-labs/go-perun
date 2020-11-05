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
	"sync/atomic"
	"unsafe"
)

// Signal is a lightweight reusable signal that can be waited on. It can be used
// to notify coroutines. Coroutines waiting for the signal will be notified once
// it is triggered, but coroutines starting to wait after triggering will not be
// notified by the same operation. This means it can be used to repeatably
// notify coroutines.
type Signal struct {
	emitter unsafe.Pointer // (*chan struct{})
	_       noCopy         // premitter this type from being copied.
}

// NewSignal creates a new signal.
func NewSignal() *Signal {
	ch := make(chan struct{})
	return &Signal{emitter: unsafe.Pointer(&ch)} // #nosec
}

// Broadcast wakes up all waiting coroutines.
func (s *Signal) Broadcast() {
	close(s.swap(make(chan struct{})))
}

// Signal wakes up a single coroutine, if any are waiting.
func (s *Signal) Signal() {
	select {
	case s.load() <- struct{}{}:
	default:
	}
}

// Wait waits until the coroutine is woken up by the signal.
func (s *Signal) Wait() {
	<-s.Done()
}

// WaitCtx waits until the context expires or the coroutine is woken up by the
// signal. Returns whether the coroutine was woken up by the signal.
func (s *Signal) WaitCtx(ctx context.Context) bool {
	select {
	case <-s.Done():
		return true
	case <-ctx.Done():
		return false
	}
}

// Done returns a channel that will be written to when the signal is next
// notified. After reading from the returned channel once, the channel should be
// discarded, and a new channel should be retrieved via a new call to Done.
func (s *Signal) Done() <-chan struct{} {
	return s.load()
}

func (s *Signal) load() chan struct{} {
	return *(*chan struct{})(atomic.LoadPointer(&s.emitter))
}

func (s *Signal) swap(next chan struct{}) (old chan struct{}) {
	nptr := unsafe.Pointer(&next) // #nosec
	return *(*chan struct{})(atomic.SwapPointer(&s.emitter, nptr))
}

// noCopy is forbidden to copy because it is a sync.Locker.
type noCopy struct{}

// Lock is a dummy to fulfill the sync.Locker interface.
func (*noCopy) Lock() {}

// Unlock is a dummy to fulfill the sync.Locker interface.
func (*noCopy) Unlock() {}
