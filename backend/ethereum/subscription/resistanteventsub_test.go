// Copyright 2021 - See NOTICE file for copyright holders.
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

package subscription_test

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/bindings/peruntoken"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/subscription"
	"perun.network/go-perun/log"
	pctx "perun.network/go-perun/pkg/context"
	pkgtest "perun.network/go-perun/pkg/test"
)

var event = func() *subscription.Event {
	return &subscription.Event{
		Name: "Approval",
		Data: new(peruntoken.PerunTokenApproval),
	}
}

// Defines a soft maximum value for the finality that tests will use.
// Must be divisible by 2 and greater than 1.
const maxFinality = 20

// TestResistantEventSub_Confirm tests that a TX is confirmed exactly
// after being included in `finalityDepth` many blocks.
func TestResistantEventSub_Confirm(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	rng := pkgtest.Prng(t)
	require := require.New(t)
	s := test.NewTokenSetup(ctx, t, rng)

	finality := rng.Int31n(maxFinality) + 1
	sub, err := subscription.Subscribe(ctx, s.CB, s.Contract, event, 0, uint64(finality))
	require.NoError(err)
	defer sub.Close()

	// Send and Confirm the TX. The simulated backend already mined a block here,
	// so the TX has 1 confirmation now.
	s.ConfirmTx(s.IncAllowance(ctx), true)
	// Wait `finality-1` blocks.
	for j := int32(0); j < finality-1; j++ {
		NoEvent(require, sub)
		s.SB.Commit()
	}
	OneEvent(require, sub)
}

// TestResistantEventSub_ReadPast tests that `ReadPast` only returns events
// that were first emitted before the sub was created even when they finalize
// after its creation.
func TestResistantEventSub_ReadPast(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	rng := pkgtest.Prng(t)
	require := require.New(t)
	s := test.NewTokenSetup(ctx, t, rng)
	numTx := 5
	finality := int(rng.Int31n(maxFinality) + 1)

	// Send some past tx.
	for i := 0; i < numTx; i++ {
		s.IncAllowance(ctx)
	}

	// Create a sub to go `numTx+finality` into the past.
	sub, err := subscription.Subscribe(ctx, s.CB, s.Contract, event, uint64(numTx), uint64(finality))
	require.NoError(err)
	defer sub.Close()

	// Finalize the past events by mining blocks.
	for i := 0; i < finality-1; i++ {
		s.SB.Commit()
	}
	// Send some future tx.
	for i := 0; i < numTx; i++ {
		s.IncAllowance(ctx)
	}
	// Finalize the future events by mining blocks.
	for i := 0; i < finality-1; i++ {
		s.SB.Commit()
	}

	// Receive exactly `numTx` past events.
	{
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		sink := make(chan *subscription.Event, numTx+1)
		err := sub.ReadPast(ctx, sink) // read only past events
		require.NoError(err)

		for i := 0; i < numTx; i++ {
			require.NotNil(<-sink)
		}

		select {
		case event := <-sink:
			require.Nil(event)
		default:
		}
	}
}

// TestResistantEventSub_ReorgConfirm tests that a TX is confirmed exactly
// after `finalityDepth` blocks even when reorgs occur.
func TestResistantEventSub_ReorgConfirm(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	rng := pkgtest.Prng(t)
	require := require.New(t)
	s := test.NewTokenSetup(ctx, t, rng)

	finality := rng.Int31n(maxFinality-2) + 2
	sub, err := subscription.Subscribe(ctx, s.CB, s.Contract, event, 0, uint64(finality))
	require.NoError(err)
	defer sub.Close()

	// Allow the reorg to go up to `maxFinality` blocks into the past.
	for i := 0; i < maxFinality; i++ {
		s.SB.Commit()
	}

	// Send and Confirm the TX
	s.IncAllowance(ctx)
	// Reorg until the block hight hits `finality` blocks.
	for h := int64(0); h < int64(finality-1); {
		d := rng.Int63n(maxFinality/2-1) + 1
		l := rng.Int63n(maxFinality/2-1) + 1
		NoEvent(require, sub)
		log.Debugf("[h=%d] Reorg with depth: %d, length: %d", h, d, l)
		s.SB.Reorg(ctx, uint64(d), func(txs []types.Transactions) []types.Transactions {
			return append(txs, make([]types.Transactions, int(d+l)-len(txs))...)
		})
		h += l
	}
	// Verify that the event arrived.
	OneEvent(require, sub)
}

// TestResistantEventSub_ReorgRemove tests that a TX is never confirmed if it
// was removed by a reorg.
func TestResistantEventSub_ReorgRemove(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	rng := pkgtest.Prng(t)
	require := require.New(t)
	s := test.NewTokenSetup(ctx, t, rng)

	finality := rng.Int31n(maxFinality-2) + 2
	sub, err := subscription.Subscribe(ctx, s.CB, s.Contract, event, 0, uint64(finality))
	require.NoError(err)
	defer sub.Close()

	// Send and Confirm the TX
	s.IncAllowance(ctx)
	NoEvent(require, sub)

	// Go back one block and remove the TX with a reorg.
	s.SB.Reorg(ctx, 1, func(txs []types.Transactions) []types.Transactions {
		return make([]types.Transactions, 2)
	})

	NoEvent(require, sub)
	// Verify that the event never arrives.
	for i := 0; i < maxFinality; i++ {
		s.SB.Commit()
	}
	time.Sleep(1 * time.Second) // give the event subscription time to catch up
	NoEvent(require, sub)
}

// TestResistantEventSub_New checks that `NewResistantEventSub` panics for
// `finalityDepth` < 1.
func TestResistantEventSub_New(t *testing.T) {
	require.PanicsWithValue(t, "finalityDepth needs to be at least 1", func() {
		subscription.NewResistantEventSub(context.Background(), nil, nil, 0)
	})
}

// NoEvent checks that no event can be read from `sub`.
func NoEvent(require *require.Assertions, sub *subscription.ResistantEventSub) {
	nEvents(require, 0, sub)
}

// OneEvent checks that exactly one event can be read from `sub`.
func OneEvent(require *require.Assertions, sub *subscription.ResistantEventSub) {
	nEvents(require, 1, sub)
}

// nEvents checks that exactly `n` events can be read from `sub`.
func nEvents(require *require.Assertions, n int, sub *subscription.ResistantEventSub) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	sink := make(chan *subscription.Event, n+1)

	err := sub.Read(ctx, sink)
	require.True(pctx.IsContextError(err))

	for i := 0; i < n; i++ {
		require.NotNil(<-sink)
	}

	select {
	case event := <-sink:
		require.Nil(event)
	default:
	}
}
