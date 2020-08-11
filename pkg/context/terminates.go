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

package context

import (
	"context"
	"time"
)

// TerminatesCtx checks whether a function terminates before a context is done.
func TerminatesCtx(ctx context.Context, fn func()) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	done := make(chan struct{}, 1)
	go func() {
		fn()
		done <- struct{}{}
	}()

	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}

// Terminates checks whether a function terminates within a certain timeout.
func Terminates(deadline time.Duration, fn func()) bool {
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()
	return TerminatesCtx(ctx, fn)
}

// TerminatesQuickly checks whether a function terminates within 20 ms.
func TerminatesQuickly(fn func()) bool {
	return Terminates(time.Millisecond*20, fn)
}
