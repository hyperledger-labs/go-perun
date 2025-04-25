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

package wire

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ctxtest "polycry.pt/poly-go/context/test"
)

// timeout controls how long to wait until we decide that something will never
// happen.
const timeout = 200 * time.Millisecond

func TestReceiver_Close(t *testing.T) {
	t.Parallel()

	r := NewReceiver()
	require.NoError(t, r.Close())
	assert.Error(t, r.Close())
}

func TestReceiver_Next(t *testing.T) {
	t.Parallel()
	e := newEnvelope(NewPingMsg())

	t.Run("Happy case", func(t *testing.T) {
		t.Parallel()
		ctxtest.AssertTerminates(t, timeout, func() {
			r := NewReceiver()
			go r.Put(e)
			re, err := r.Next(context.Background())
			require.NoError(t, err)
			assert.Same(t, e, re)
		})
	})

	t.Run("Closed before", func(t *testing.T) {
		t.Parallel()
		ctxtest.AssertTerminates(t, timeout, func() {
			r := NewReceiver()
			r.Close()
			re, err := r.Next(context.Background())
			assert.Nil(t, re)
			assert.Error(t, err)
		})
	})

	t.Run("Delayed close", func(t *testing.T) {
		t.Parallel()
		ctxtest.AssertTerminates(t, timeout*2, func() {
			r := NewReceiver()
			go func() {
				time.Sleep(timeout)
				r.Close()
			}()
			re, err := r.Next(context.Background())
			assert.Nil(t, re)
			assert.Error(t, err)
		})
	})

	t.Run("Context instant timeout", func(t *testing.T) {
		t.Parallel()
		ctxtest.AssertTerminates(t, timeout, func() {
			r := NewReceiver()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			re, err := r.Next(ctx)
			assert.Nil(t, re)
			assert.Error(t, err)
		})
	})

	t.Run("Context delayed timeout", func(t *testing.T) {
		t.Parallel()
		ctxtest.AssertTerminates(t, timeout*2, func() {
			r := NewReceiver()
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			re, err := r.Next(ctx)
			assert.Nil(t, re)
			assert.Error(t, err)
		})
	})
}
