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

package client

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
)

// AdjudicatorEventHandler represents an interface for handling adjudicator events.
type AdjudicatorEventHandler interface {
	HandleAdjudicatorEvent(channel.AdjudicatorEvent)
}

// Watch watches the adjudicator for channel events and responds accordingly.
// The handler is notified about the corresponding events.
//
// The routine takes care that if an old state is registered, the on-chain state
// is refuted with the most recent event available. In such a case, the handler
// may receive multiple registered events in short succession.
//
// Returns TxTimedoutError when watcher refutes with the most recent state and
// the program times out waiting for a transaction to be mined.
// Returns ChainNotReachableError if the connection to the blockchain network
// fails when sending a transaction to / reading from the blockchain.
func (c *Channel) Watch(h AdjudicatorEventHandler) error {
	log := c.Log().WithField("proc", "watcher")
	defer log.Info("Watcher returned.")

	// Subscribe to state changes
	ctx := c.Ctx()
	sub, err := c.adjudicator.Subscribe(ctx, c.Params())
	if err != nil {
		return errors.WithMessage(err, "subscribing to adjudicator state changes")
	}
	// nolint:errcheck
	defer sub.Close()
	// nolint:errcheck,gosec
	c.OnCloseAlways(func() { sub.Close() })

	// Wait for state changed event
	for e := sub.Next(); e != nil; e = sub.Next() {
		log.Infof("event %v", e)

		// Update machine phase
		if err := c.setMachinePhase(ctx, e); err != nil {
			return errors.WithMessage(err, "setting machine phase")
		}

		// Special handling of RegisteredEvent
		if e, ok := e.(*channel.RegisteredEvent); ok {
			// Assert backend version not greater than local version.
			if e.Version() > c.State().Version {
				// If the implementation works as intended, this should never happen.
				log.Panicf("watch: registered: expected version less than or equal to %d, got version %d", c.machine.State().Version, e.Version)
			}

			// If local version greater than backend version, register local state.
			if e.Version() < c.State().Version {
				if err := c.Register(ctx); err != nil {
					return errors.WithMessage(err, "registering")
				}
			}
		}

		// Notify handler
		go h.HandleAdjudicatorEvent(e)
	}

	err = sub.Err()
	log.Debugf("Subscription closed: %v", err)
	return errors.WithMessage(err, "subscription closed")
}

// Register registers the channel on the adjudicator.
//
// Returns TxTimedoutError when the program times out waiting for a transaction
// to be mined.
// Returns ChainNotReachableError if the connection to the blockchain network
// fails when sending a transaction to / reading from the blockchain.
func (c *Channel) Register(ctx context.Context) error {
	// Lock channel machine.
	if !c.machMtx.TryLockCtx(ctx) {
		return errors.WithMessage(ctx.Err(), "locking machine")
	}
	defer c.machMtx.Unlock()

	return c.register(ctx)
}

// ProgressBy progresses the channel state in the adjudicator backend.
//
// Returns TxTimedoutError when the program times out waiting for a transaction
// to be mined.
// Returns ChainNotReachableError if the connection to the blockchain network
// fails when sending a transaction to / reading from the blockchain.
func (c *Channel) ProgressBy(ctx context.Context, update func(*channel.State)) error {
	// Lock machine
	if !c.machMtx.TryLockCtx(ctx) {
		return errors.Errorf("locking machine mutex in time: %v", ctx.Err())
	}
	defer c.machMtx.Unlock()

	// Store current state
	ar := c.machine.AdjudicatorReq()

	// Update state
	state := c.machine.State().Clone()
	state.Version++
	update(state)

	// Apply state in machine and generate signature
	if err := c.machine.SetProgressing(ctx, state); err != nil {
		return errors.WithMessage(err, "updating machine")
	}
	sig, err := c.machine.Sig(ctx)
	if err != nil {
		return errors.WithMessage(err, "signing")
	}

	// Create and send request
	pr := channel.NewProgressReq(ar, state, sig)
	return errors.WithMessage(c.adjudicator.Progress(ctx, *pr), "progressing")
}

// Settle concludes a channel and withdraws the funds.
//
// Returns TxTimedoutError when the program times out waiting for a transaction
// to be mined.
// Returns ChainNotReachableError if the connection to the blockchain network
// fails when sending a transaction to / reading from the blockchain.
func (c *Channel) Settle(ctx context.Context, secondary bool) error {
	return c.SettleWithSubchannels(ctx, nil, secondary)
}

// SettleWithSubchannels concludes a channel and withdraws the funds.
//
// If the channel is a ledger channel with locked funds, additionally subStates
// can be supplied to also conclude the corresponding sub-channels.
//
// Returns TxTimedoutError when the program times out waiting for a transaction
// to be mined.
// Returns ChainNotReachableError if the connection to the blockchain network
// fails when sending a transaction to / reading from the blockchain.
func (c *Channel) SettleWithSubchannels(ctx context.Context, subStates channel.StateMap, secondary bool) error {
	// Lock channel machine.
	if !c.machMtx.TryLockCtx(ctx) {
		return errors.WithMessage(ctx.Err(), "locking machine")
	}
	defer c.machMtx.Unlock()

	if err := c.machine.SetWithdrawing(ctx); err != nil {
		return errors.WithMessage(err, "setting machine to withdrawing phase")
	}

	switch {
	case c.IsLedgerChannel():
		req := c.machine.AdjudicatorReq()
		req.Secondary = secondary
		if err := c.adjudicator.Withdraw(ctx, req, subStates); err != nil {
			return errors.WithMessage(err, "calling Withdraw")
		}
	case c.IsSubChannel():
		if c.hasLockedFunds() {
			return errors.New("cannot settle off-chain with locked funds")
		}
		if err := c.withdrawIntoParent(ctx); err != nil {
			return errors.WithMessage(err, "withdrawing into parent channel")
		}
	default:
		panic("invalid channel type")
	}

	if err := c.machine.SetWithdrawn(ctx); err != nil {
		return errors.WithMessage(err, "setting machine phase")
	}

	c.Log().Info("Withdrawal successful.")
	c.wallet.DecrementUsage(c.machine.Account().Address())
	return nil
}

// register calls Register on the adjudicator with the current channel state and
// progresses the machine phases.
//
// The caller is expected to have locked the channel mutex.
func (c *Channel) register(ctx context.Context) error {
	if err := c.machine.SetRegistering(ctx); err != nil {
		return errors.WithMessage(err, "setting the machine phase")
	}

	if err := c.adjudicator.Register(ctx, c.machine.AdjudicatorReq()); err != nil {
		return errors.WithMessage(err, "calling Register")
	}

	return c.machine.SetRegistered(ctx)
}

func (c *Channel) setMachinePhase(ctx context.Context, e channel.AdjudicatorEvent) (err error) {
	// Lock machine
	if !c.machMtx.TryLockCtx(ctx) {
		return errors.WithMessage(ctx.Err(), "locking machine")
	}
	defer c.machMtx.Unlock()

	switch e := e.(type) {
	case *channel.RegisteredEvent:
		err = c.machine.SetRegistered(ctx)
	case *channel.ProgressedEvent:
		err = c.machine.SetProgressed(ctx, e)
	case *channel.ConcludedEvent:
		// Do nothing as there is currently no corresponding phase in the channel machine.
	default:
		c.Log().Panic("unsupported event type")
	}

	return
}
