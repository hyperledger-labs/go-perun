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
	"perun.network/go-perun/watcher"
	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/sync"
)

// AdjudicatorEventHandler represents an interface for handling adjudicator events.
type AdjudicatorEventHandler interface {
	HandleAdjudicatorEvent(channel.AdjudicatorEvent)
}

// Watch registers the channel with the watcher, which watches for channel
// events and responds accordingly. The handler is notified about the
// corresponding events.
//
// The watcher takes care that if an old state is registered, the on-chain
// state is refuted by registering the channel tree with the most recent states
// available to the watcher. In such a case, the handler may receive multiple
// registered events in short succession.
//
// It should be started as a go-routine and returns when the channel is closed.
func (c *Channel) Watch(h AdjudicatorEventHandler) error {
	log := c.Log().WithField("proc", "watcher")
	defer log.Info("Watcher returned.")

	statesPub, eventsSub, err := c.startWatching()
	if err != nil {
		return err
	}
	c.machMtx.Lock()
	c.statesPub = statesPub
	c.machMtx.Unlock()
	if err := c.handleEvents(eventsSub, h); err != nil {
		return errors.WithMessage(err, "handling events from watcher")
	}

	err = errors.WithMessage(eventsSub.Err(), "subscription closed")
	log.Debugf("Subscription closed: %v", err)
	return err
}

func (c *Channel) startWatching() (watcher.StatesPub, watcher.AdjudicatorSub, error) {
	c.machMtx.Lock()
	defer c.machMtx.Unlock()

	currentTx := c.machine.CurrentTX()
	signedState := channel.SignedState{
		Params: c.Params(),
		State:  currentTx.State.Clone(),
		Sigs:   currentTx.Sigs,
	}

	statesPub, eventsSub, err := func() (watcher.StatesPub, watcher.AdjudicatorSub, error) {
		if c.IsLedgerChannel() {
			return c.client.watcher.StartWatchingLedgerChannel(c.Ctx(), signedState)
		}
		return c.client.watcher.StartWatchingSubChannel(c.Ctx(), c.parent.ID(), signedState)
	}()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "registering channel with the watcher")
	}
	ok := c.OnCloseAlways(func() {
		err := c.client.watcher.StopWatching(c.Ctx(), c.ID())
		if err != nil {
			c.Log().Errorf("Error de-registering channel from watcher: %v", err)
		}
	})
	if !ok {
		return nil, nil, errors.WithMessage(err, "channel already closed")
	}
	return statesPub, eventsSub, nil
}

func (c *Channel) handleEvents(eventsSub watcher.AdjudicatorSub, h AdjudicatorEventHandler) error {
	for {
		select {
		case e, ok := <-eventsSub.EventStream():
			if !ok {
				return nil
			}
			log.WithField("channel", c.Params().ID()).WithField("participant", c.Idx()).Infof("event %T: %v", e, e)
			if err := c.setMachinePhase(c.Ctx(), e); err != nil {
				return errors.WithMessage(err, "setting machine phase")
			}

			// Notify handler
			go h.HandleAdjudicatorEvent(e)
		case <-c.Ctx().Done():
			c.Log().Info("Exiting watcher as channel context was cancelled")
			return nil
		}
	}
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

// registerDispute registers a dispute for the channel and all its relatives.
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

	err = c.adjudicator.Register(ctx, c.machine.AdjudicatorReq(), subStates)
	if err != nil {
		return errors.WithMessage(err, "calling Register")
	}

	err = c.setRegisteredRecursive(ctx)
	if err != nil {
		return errors.WithMessage(err, "setting phase `Registered` recursive")
	}

	return nil
}

// ForceUpdate enforces a state update through the adjudicator as specified by
// the `updater` function.
//
// The updater function must adhere to the following rules:
// - The version must not be changed. It is incremented automatically.
// - The assets must not be changed.
// - The locked allocation must not be changed.
//
// Returns TxTimedoutError when the program times out waiting for a transaction
// to be mined. Returns ChainNotReachableError if the connection to the
// blockchain network fails when sending a transaction to / reading from the
// blockchain.
func (c *Channel) ForceUpdate(ctx context.Context, updater func(*channel.State)) error {
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
	state := ar.Tx.State.Clone()
	updater(state)
	state.Version++

	// Check state transition.
	if err := c.machine.ValidTransition(state); err != nil {
		return errors.WithMessage(err, "validating state transition")
	}

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

// Settle concludes the channel and withdraws the funds.
//
// This only works if the channel is concludable.
// - A ledger channel is concludable, if it is final or if it has been disputed
// before and the dispute timeout has passed.
// - Sub-channels and virtual channels are only concludable if they are final
// and do not have any sub-channels. Otherwise, this means a dispute has
// occurred and the corresponding ledger channel must be disputed.
//
// Returns TxTimedoutError when the program times out waiting for a transaction
// to be mined.
// Returns ChainNotReachableError if the connection to the blockchain network
// fails when sending a transaction to / reading from the blockchain.
func (c *Channel) Settle(ctx context.Context, secondary bool) (err error) {
	if !c.State().IsFinal {
		err := c.ensureRegistered(ctx)
		if err != nil {
			return err
		}
	}

	// Lock machines of channel and all subchannels recursively.
	l, err := c.tryLockRecursive(ctx)
	defer l.Unlock()
	if err != nil {
		return errors.WithMessage(err, "locking recursive")
	}

	// Set phase `Withdrawing`.
	if err = c.applyRecursive(func(c *Channel) error {
		if c.machine.Phase() == channel.Withdrawn {
			return nil
		}
		return c.machine.SetWithdrawing(ctx)
	}); err != nil {
		return errors.WithMessage(err, "setting phase `Withdrawing` recursive")
	}

	// Withdraw.
	err = c.withdraw(ctx, secondary)
	if err != nil {
		return
	}

	// Set phase `Withdrawn`.
	if err = c.applyRecursive(func(c *Channel) error {
		// Skip if already withdrawn.
		if c.machine.Phase() == channel.Withdrawn {
			return nil
		}
		return c.machine.SetWithdrawn(ctx)
	}); err != nil {
		return errors.WithMessage(err, "setting phase `Withdrawn` recursive")
	}

	// Decrement account usage.
	if err = c.applyRecursive(func(c *Channel) (err error) {
		// Skip if we are not a participant, e.g., if this is a virtual channel and we are the hub.
		if c.IsVirtualChannel() {
			ourID := c.parent.Peers()[c.parent.Idx()]
			if !c.hasParticipant(ourID) {
				return
			}
		}
		c.wallet.DecrementUsage(c.machine.Account().Address())
		return
	}); err != nil {
		return errors.WithMessage(err, "decrementing account usage")
	}

	c.Log().Info("Withdrawal successful.")
	return nil
}

func (c *Channel) withdraw(ctx context.Context, secondary bool) error {
	switch {
	case c.IsLedgerChannel():
		subStates, err := c.subChannelStateMap()
		if err != nil {
			return errors.WithMessage(err, "creating sub-channel state map")
		}
		req := c.machine.AdjudicatorReq()
		req.Secondary = secondary
		if err := c.adjudicator.Withdraw(ctx, req, subStates); err != nil {
			return errors.WithMessage(err, "calling Withdraw")
		}

	case c.IsSubChannel():
		if c.hasLockedFunds() {
			return errors.New("cannot settle off-chain with locked funds")
		}
		if err := c.withdrawSubChannelIntoParent(ctx); err != nil {
			return errors.WithMessage(err, "withdrawing into parent channel")
		}

	case c.IsVirtualChannel():
		if c.hasLockedFunds() {
			return errors.New("cannot settle off-chain with locked funds")
		}
		if err := c.parent.withdrawVirtualChannel(ctx, c); err != nil {
			return errors.WithMessage(err, "withdrawing into parent channel")
		}

	default:
		panic("invalid channel type")
	}
	return nil
}

// hasParticipant returns we are participating in the channel.
func (c *Channel) hasParticipant(id wire.Address) bool {
	for _, p := range c.Peers() {
		if id.Equal(p) {
			return true
		}
	}
	return false
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
	err = c.applyRecursive(func(c *Channel) error {
		if !c.machMtx.TryLockCtx(ctx) {
			return errors.Errorf("locking machine mutex in time: %v", ctx.Err())
		}
		l = append(l, &c.machMtx)
		return nil
	})
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

// applyRecursive applies the function to the channel and its sub-channels recursively.
func (c *Channel) applyRecursive(f func(*Channel) error) (err error) {
	err = f(c)
	if err != nil {
		return err
	}

	err = c.applyToSubChannelsRecursive(f)
	return
}

// setRegisteringRecursive sets the machine phase of the channel and all of its sub-channels to `Registering`.
// Assumes that the channel machine has been locked.
func (c *Channel) setRegisteringRecursive(ctx context.Context) (err error) {
	return c.applyRecursive(func(c *Channel) error {
		return c.machine.SetRegistering(ctx)
	})
}

// setRegisteredRecursive sets the machine phase of the channel and all of its sub-channels to `Registered`.
// Assumes that the channel machine has been locked.
func (c *Channel) setRegisteredRecursive(ctx context.Context) (err error) {
	return c.applyRecursive(func(c *Channel) error {
		return c.machine.SetRegistered(ctx)
	})
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

// subChannelStateMap gathers the state of all sub-channels recursively.
// Assumes sub-channels are locked.
func (c *Channel) subChannelStateMap() (states channel.StateMap, err error) {
	states = channel.MakeStateMap()
	err = c.applyToSubChannelsRecursive(func(c *Channel) error {
		states[c.ID()] = c.state()
		return nil
	})
	return
}

// ensureRegistered ensures that the channel is registered.
func (c *Channel) ensureRegistered(ctx context.Context) error {
	phase := c.Phase()
	if phase == channel.Registered ||
		phase == channel.Progressing ||
		phase == channel.Progressed {
		return nil
	}

	registered := make(chan error)
	go func() {
		registered <- c.awaitRegistered(ctx)
	}()

	// Register.
	err := c.registerDispute(ctx)
	if err != nil {
		// Only log because channel may already be registered.
		c.Log().Warnf("registering: %v", err)
	}

	select {
	case err = <-registered:
	case <-ctx.Done():
		err = ctx.Err()
	}
	return err
}

// awaitRegistered scans for an event indicating that the channel has been
// registered and waits for the event timeout to elapse.
func (c *Channel) awaitRegistered(ctx context.Context) error {
	// Start event subscription.
	sub, err := c.adjudicator.Subscribe(ctx, c.Params().ID())
	if err != nil {
		return errors.WithMessage(err, "subscribing to adjudicator events")
	}
	defer func() {
		if err := sub.Close(); err != nil {
			c.Log().Warn("Subscription closed with error:", err)
		}
	}()

	// Scan for event.
	for e := sub.Next(); e != nil; e = sub.Next() {
		switch e.(type) {
		case *channel.RegisteredEvent:
		case *channel.ProgressedEvent:
		case *channel.ConcludedEvent:
		default:
			log.Warnf("unrecognized event type: %T", e)
			continue
		}

		// We set phase registered to ensure that we are in the correct phase
		// even if the channel has been registered by someone else.
		l, err := c.tryLockRecursive(ctx)
		defer l.Unlock()
		if err != nil {
			return errors.WithMessage(err, "locking recursive")
		}
		err = c.setRegisteredRecursive(ctx)
		if err != nil {
			return errors.WithMessage(err, "setting phase `Registered` recursive")
		}

		// Wait until end of channel phase.
		return e.Timeout().Wait(ctx)
	}
	return sub.Err()
}
