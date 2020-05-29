// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

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

	// We use nil contexts in the following because we're working with a simulated
	// blockchain, which ignores the ctx.
	for i := 0; i < 10; i++ {
		assert.False(bt.IsElapsed(nil))
		sb.Commit() // advances block time by 10 sec
	}
	assert.True(bt.IsElapsed(nil))
}

func TestBlockTimeout_Wait(t *testing.T) {
	sb := test.NewSimulatedBackend()
	bt := ethchannel.NewBlockTimeout(sb, 100)

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
		case <-time.After(100 * time.Millisecond):
			t.Error("expected Wait to return")
		}
	})

	t.Run("normalWait", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		wait := make(chan error)
		go func() {
			wait <- bt.Wait(ctx)
		}()

		for i := 0; i < 10; i++ {
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
		case <-time.After(100 * time.Millisecond):
			t.Error("expected Wait to return after timeout is reached")
		}
	})
}
