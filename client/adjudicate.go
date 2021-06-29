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
	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/sync"
)

// AdjudicatorEventHandler represents an interface for handling adjudicator events.
type AdjudicatorEventHandler interface {
	HandleAdjudicatorEvent(channel.AdjudicatorEvent)
}

// Watch watches the adjudicator for channel events and responds accordingly.
// The handler is notified about the corresponding events.
//
// The routine takes care that if an old state is registered, the on-chain state
// is refuted with the most recent event available by registering the channel
// tree. In such a case, the handler may receive multiple registered events in
// short succession.
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
				if err := c.registerDispute(ctx); err != nil {
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

// registerDispute registers the channel and all its relatives on the adjudicator.
//
// Returns TxTimedoutError when the program times out waiting for a transaction
// to be mined.
// Returns ChainNotReachableError if the connection to the blockchain network
// fails when sending a transaction to / reading from the blockchain.
func (c *Channel) registerDispute(ctx context.Context) error {
	// If this is not the root, go up one level.
	// Once we are at the root, we register the whole channel tree together.
	if c.parent != nil {
		return c.parent.registerDispute(ctx)
	}

	// Lock machines of channel and all subchannels recursively.
	l, err := c.tryLockRecursive(ctx)
	defer l.Unlock()
	if err != nil {
		return errors.WithMessage(err, "locking recursive")
	}

	err = c.setRegisteringRecursive(ctx)
	if err != nil {
		return errors.WithMessage(err, "setting phase `Registering` recursive")
	}

	subStates, err := c.gatherSubChannelStates()
	if err != nil {
		return errors.WithMessage(err, "gathering sub-channel states")
	}

	err = c.adjudicator.Register(ctx, channel.RegisterReq{
		AdjudicatorReq: c.machine.AdjudicatorReq(),
		SubChannels:    subStates,
	})
	if err != nil {
		return errors.WithMessage(err, "calling Register")
	}

	err = c.setRegisteredRecursive(ctx)
	if err != nil {
		return errors.WithMessage(err, "setting phase `Registered` recursive")
	}

	return nil
}

// ForceUpdate progresses the channel state in the adjudicator backend.
//
// Returns TxTimedoutError when the program times out waiting for a transaction
// to be mined.
// Returns ChainNotReachableError if the connection to the blockchain network
// fails when sending a transaction to / reading from the blockchain.
func (c *Channel) ForceUpdate(ctx context.Context, update func(*channel.State)) error {
	err := c.ensureRegistered(ctx)
	if err != nil {
		return err
	}

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

func (c *Channel) ensureRegistered(ctx context.Context) (err error) {
	phase := c.Phase()
	if phase == channel.Registered || phase == channel.Progressed {
		return
	}

	registeredEvents := make(chan *channel.RegisteredEvent)

	// Start event subscription.
	sub, err := c.adjudicator.Subscribe(ctx, c.Params())
	if err != nil {
		return errors.WithMessage(err, "subscribing to adjudicator events")
	}
	go waitForRegisteredEvent(sub, registeredEvents)

	// Register.
	err = c.registerDispute(ctx)
	if err != nil {
		return errors.WithMessage(err, "registering")
	}

	select {
	case e := <-registeredEvents:
		err = e.Timeout().Wait(ctx)
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	return
}

func waitForRegisteredEvent(sub channel.AdjudicatorSubscription, events chan<- *channel.RegisteredEvent) {
	defer func() {
		err := sub.Close()
		if err != nil {
			log.Warn("Subscription closed with error:", err)
		}
	}()

	// Scan for registered event.
	for e := sub.Next(); e != nil; e = sub.Next() {
		if e, ok := e.(*channel.RegisteredEvent); ok {
			events <- e
			return
		}
	}
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
		if err := c.adjudicator.Withdraw(ctx, channel.WithdrawReq{
			AdjudicatorReq: req,
			SubChannels:    subStates,
		}); err != nil {
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

type mutexList []*sync.Mutex

func (a mutexList) Unlock() {
	for _, m := range a {
		m.Unlock()
	}
}

// tryLockRecursive tries to lock the channel and all of its sub-channels.
// It returns a list of all the mutexes that have been locked.
func (c *Channel) tryLockRecursive(ctx context.Context) (l mutexList, err error) {
	l = mutexList{}
	f := func(c *Channel) error {
		if !c.machMtx.TryLockCtx(ctx) {
			return errors.Errorf("locking machine mutex in time: %v", ctx.Err())
		}
		l = append(l, &c.machMtx)
		return nil
	}

	err = f(c)
	if err != nil {
		return
	}

	err = c.applyToSubChannelsRecursive(f)
	return
}

// applyToSubChannelsRecursive applies the function to all sub-channels recursively.
func (c *Channel) applyToSubChannelsRecursive(f func(*Channel) error) (err error) {
	for _, subAlloc := range c.state().Locked {
		subID := subAlloc.ID
		var subCh *Channel
		subCh, err = c.client.Channel(subID)
		if err != nil {
			err = errors.WithMessagef(err, "getting sub-channel: %v", subID)
			return
		}
		err = f(subCh)
		if err != nil {
			return
		}
		err = subCh.applyToSubChannelsRecursive(f)
		if err != nil {
			return
		}
	}
	return
}

// setRegisteringRecursive sets the machine phase of the channel and all of its sub-channels to `Registering`.
// Assumes that the channel machine has been locked.
func (c *Channel) setRegisteringRecursive(ctx context.Context) (err error) {
	f := func(c *Channel) error {
		return c.machine.SetRegistering(ctx)
	}

	err = f(c)
	if err != nil {
		return err
	}

	err = c.applyToSubChannelsRecursive(f)
	return
}

// setRegisteredRecursive sets the machine phase of the channel and all of its sub-channels to `Registered`.
// Assumes that the channel machine has been locked.
func (c *Channel) setRegisteredRecursive(ctx context.Context) (err error) {
	f := func(c *Channel) error {
		return c.machine.SetRegistered(ctx)
	}

	err = f(c)
	if err != nil {
		return err
	}

	err = c.applyToSubChannelsRecursive(f)
	return
}

// gatherSubChannelStates gathers the state of all sub-channels recursively.
// Assumes sub-channels are locked.
func (c *Channel) gatherSubChannelStates() (states []channel.SignedState, err error) {
	states = []channel.SignedState{}
	err = c.applyToSubChannelsRecursive(func(c *Channel) error {
		states = append(states, channel.SignedState{
			Params: c.Params(),
			State:  c.machine.CurrentTX().State,
			Sigs:   c.machine.CurrentTX().Sigs,
		})
		return nil
	})
	return
}
