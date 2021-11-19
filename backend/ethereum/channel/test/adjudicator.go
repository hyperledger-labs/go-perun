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

package test

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
)

// A SimAdjudicator is an Adjudicator for simulated backends. Its Register
// method and subscription return a *channel.RegisteredEvent whose Timeout is a
// SimTimeout. SimTimeouts advance the clock of the simulated backend when Wait
// is called.
type SimAdjudicator struct {
	ethchannel.Adjudicator
	sb *SimulatedBackend
}

// NewSimAdjudicator returns a new SimAdjudicator for the given backend. The
// backend must be a SimulatedBackend or it panics.
func NewSimAdjudicator(backend ethchannel.ContractBackend, contract common.Address, receiver common.Address, acc accounts.Account) *SimAdjudicator {
	sb, ok := backend.ContractInterface.(*SimulatedBackend)
	if !ok {
		panic("SimAdjudicator can only be created with a SimulatedBackend.")
	}
	return &SimAdjudicator{
		Adjudicator: *ethchannel.NewAdjudicator(backend, contract, receiver, acc),
		sb:          sb,
	}
}

// Subscribe returns a RegisteredEvent subscription on the simulated
// blockchain backend.
func (a *SimAdjudicator) Subscribe(ctx context.Context, chID channel.ID) (channel.AdjudicatorSubscription, error) {
	sub, err := a.Adjudicator.Subscribe(ctx, chID)
	if err != nil {
		return nil, err
	}
	return &SimRegisteredSub{
		RegisteredSub: sub.(*ethchannel.RegisteredSub),
		sb:            a.sb,
	}, nil
}

// A SimRegisteredSub embeds an ethereum/channel.RegisteredSub, converting
// normal TimeTimeouts to SimTimeouts.
type SimRegisteredSub struct {
	*ethchannel.RegisteredSub
	sb *SimulatedBackend
}

// Next calls Next on the underlying subscription, converting the TimeTimeout to
// a SimTimeout.
func (r *SimRegisteredSub) Next() channel.AdjudicatorEvent {
	switch ev := r.RegisteredSub.Next().(type) {
	case nil:
		return nil
	case *channel.RegisteredEvent:
		if ev == nil {
			return nil
		}
		ev.TimeoutV = block2SimTimeout(r.sb, ev.Timeout().(*ethchannel.BlockTimeout))
		return ev
	case *channel.ProgressedEvent:
		if ev == nil {
			return nil
		}
		ev.TimeoutV = block2SimTimeout(r.sb, ev.Timeout().(*ethchannel.BlockTimeout))
		return ev
	case *channel.ConcludedEvent:
		if ev == nil {
			return nil
		}
		ev.TimeoutV = block2SimTimeout(r.sb, ev.Timeout().(*ethchannel.BlockTimeout))
		return ev
	default:
		log.Panicf("unknown AdjudicatorEvent type: %t", ev)
		return nil // never reached
	}
}

func block2SimTimeout(sb *SimulatedBackend, t *ethchannel.BlockTimeout) *SimTimeout {
	return &SimTimeout{t.Time, sb}
}

// A SimTimeout is a timeout on a simulated blockchain. The first call to Wait
// advances the clock of the simulated blockchain past the timeout. Access to
// the blockchain by different SimTimeouts is guarded by a shared mutex.
type SimTimeout struct {
	Time uint64
	sb   *SimulatedBackend
}

// IsElapsed returns whether the timeout is higher than the current block's
// timestamp.
// Access to the blockchain by different SimTimeouts is guarded by a shared mutex.
func (t *SimTimeout) IsElapsed(ctx context.Context) bool {
	if !t.sb.clockMu.TryLockCtx(ctx) {
		return false // subsequent Wait call will expose error to caller
	}
	defer t.sb.clockMu.Unlock()

	return t.timeLeft(ctx) <= 0
}

// Wait advances the clock of the simulated blockchain past the timeout.
// Access to the blockchain by different SimTimeouts is guarded by a shared mutex.
func (t *SimTimeout) Wait(ctx context.Context) error {
	if !t.sb.clockMu.TryLockCtx(ctx) {
		return errors.New("clock mutex could not be locked")
	}
	defer t.sb.clockMu.Unlock()

	if d := t.timeLeft(ctx); d > 0 {
		if err := t.sb.AdjustTime(time.Duration(d) * time.Second); err != nil {
			return errors.Wrap(err, "adjusting time")
		}
		t.sb.Commit()
	}
	return nil
}

func (t *SimTimeout) timeLeft(ctx context.Context) int64 {
	// context is ignored by sim blockchain anyways
	h, err := t.sb.HeaderByNumber(ctx, nil)
	if err != nil { // should never happen with a sim blockchain
		panic(fmt.Sprint("Error getting latest block: ", err))
	}
	return int64(t.Time) - int64(h.Time)
}

// String returns the timeout in absolute seconds as a string.
func (t *SimTimeout) String() string {
	return fmt.Sprintf("<Sim timeout: %v>", t.Time)
}
