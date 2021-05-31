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

package channel_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
)

func TestBlockTimeout_IsElapsed(t *testing.T) {
	assert := assert.New(t)
	sb := test.NewSimulatedBackend()
	bt := ethchannel.NewBlockTimeout(sb, 100)

	// We use context.TODO() in the following because we're working with a simulated
	// blockchain, which ignores the ctx.
	for i := 0; i < 10; i++ {
		assert.False(bt.IsElapsed(context.TODO()))
		sb.Commit() // advances block time by 10 sec
	}
	assert.True(bt.IsElapsed(context.TODO()))
}

func TestBlockTimeout_Wait(t *testing.T) {
	const (
		ctxTimeout   = 1 * time.Second   // context timeout per commit
		blockTimeout = 100               // in sec
		numBlocks    = blockTimeout / 10 // Commit() advances by 10 sec
	)

	sb := test.NewSimulatedBackend()
	bt := ethchannel.NewBlockTimeout(sb, blockTimeout)

	t.Run("cancelWait", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		wait := make(chan error)
		go func() {
			wait <- bt.Wait(ctx)
		}()

		cancel()
		select {
		case err := <-wait:
			assert.Error(t, err)
		case <-time.After(ctxTimeout):
			t.Error("expected Wait to return")
		}
	})

	t.Run("normalWait", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*ctxTimeout)
		defer cancel()
		wait := make(chan error)
		go func() {
			wait <- bt.Wait(ctx)
		}()

		// Wait for the go-routine above to start.
		time.Sleep(100 * time.Millisecond)
		for i := 0; i < numBlocks; i++ {
			select {
			case err := <-wait:
				t.Error("Wait returned before timeout with error", err)
			default: // Wait shouldn't return before the timeout is reached
			}
			sb.Commit() // advances block time by 10 sec
		}
		select {
		case err := <-wait:
			assert.NoError(t, err)
		case <-time.After(10 * ctxTimeout):
			t.Error("expected Wait to return after timeout is reached")
		}
	})
}
