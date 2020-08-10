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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const timeout = 200 * time.Millisecond

func TestTerminatesCtx(t *testing.T) {
	// Test often, to detect if there are rare execution branches (due to
	// 'select' statements).
	for i := 0; i < 256; i++ {
		t.Run("immediate deadline", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			assert.False(t, TerminatesCtx(ctx, func() {}))
		})
	}

	t.Run("delayed deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		cancel()
		assert.False(t, TerminatesCtx(ctx, func() {
			<-time.After(2 * timeout)
		}))
	})

	t.Run("no deadline", func(t *testing.T) {
		assert.True(t, TerminatesCtx(context.Background(), func() {}))
	})
}

func TestTerminates(t *testing.T) {
	// Test often, to detect if there are rare execution branches (due to
	// 'select' statements).
	for i := 0; i < 256; i++ {
		t.Run("immediate deadline", func(t *testing.T) {
			assert.False(t, Terminates(-1, func() { <-time.After(time.Second) }))
		})
	}

	t.Run("delayed deadline", func(t *testing.T) {
		assert.False(t, Terminates(timeout, func() {
			<-time.After(2 * timeout)
		}))

		assert.True(t, Terminates(2*timeout, func() {
			<-time.After(timeout)
		}))
	})
}
