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

package sync_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/context/test"
	"perun.network/go-perun/pkg/sync"
)

const timeout = 100 * time.Millisecond

func TestCloser_Closed(t *testing.T) {
	t.Parallel()
	var c sync.Closer

	assert.NotNil(t, c.Closed())
	select {
	case _, ok := <-c.Closed():
		t.Fatalf("Closed() should not yield a value, ok = %t", ok)
	default:
	}

	require.NoError(t, c.Close())

	test.AssertTerminates(t, timeout, func() {
		_, ok := <-c.Closed()
		assert.False(t, ok)
	})
}

func TestCloser_Ctx(t *testing.T) {
	t.Parallel()
	var c sync.Closer
	ctx := c.Ctx()
	assert.NoError(t, ctx.Err())
	assert.Nil(t, ctx.Value(nil))
	_, ok := ctx.Deadline()
	assert.False(t, ok)

	select {
	case <-ctx.Done():
		t.Error("context should not be closed")
	default: // expected
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		<-ctx.Done()
		assert.Same(t, ctx.Err(), context.Canceled)
	}()
	assert.NoError(t, c.Close())
	<-done
}
