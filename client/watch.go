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
	"time"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
)

const waitEventDuration = 1 * time.Second

// Watch starts the channel watcher routine. It subscribes to RegisteredEvents
// on the adjudicator. If an event is registered, it is handled by making sure
// the latest state is registered and then all funds withdrawn to the receiver
// specified in the adjudicator that was passed to the channel.
//
// If handling failed, the watcher routine returns the respective error. It is
// the user's job to restart the watcher after the cause of the error got fixed.
func (c *Channel) Watch() error {
	log := c.Log().WithField("proc", "watcher")
	defer log.Info("Watcher returned.")

	ctx := c.Ctx()
	sub, err := c.adjudicator.Subscribe(ctx, c.Params())
	if err != nil {
		return errors.WithMessage(err, "subscribing to RegisteredEvents")
	}
	closeSub := func() {
		if err := sub.Close; err != nil {
			c.Log().Warnf("Error closing adjudicator subscription: %v", err)
		}
	}
	defer closeSub()
	c.OnCloseAlways(closeSub)

	// Wait for on-chain event
	event := sub.Next()
	log.Infof("New AdjudicatorEvent: %v", event)
	switch ev := event.(type) {
	case nil:
		err := sub.Err() // err might be nil if subscription got orderly closed
		log.Debugf("Subscription closed: %v", err)
		return errors.WithMessage(err, "subscription closed")
	case *channel.RegisteredEvent:
		return errors.WithMessage(
			c.handleRegisteredEvent(ctx, ev),
			"handling RegisteredEvent")
	case *channel.ProgressedEvent:
		c.Log().Panic("Progressed event handling not implemented yet")
	}

	return errors.New("unexpect return")
}

// handleRegisteredEvent stores the passed RegisteredEvent to the machine and
// settles the channel.
func (c *Channel) handleRegisteredEvent(ctx context.Context, reg *channel.RegisteredEvent) error {
	log := c.Log().WithField("proc", "watcher")
	// Lock machine while registering is in progress.
	if !c.machMtx.TryLockCtx(ctx) {
		return errors.Errorf("locking machine mutex in time: %v", ctx.Err())
	}
	defer c.machMtx.Unlock()

	if c.machine.Phase() == channel.Withdrawn {
		// If a Settle call by the user caused this event, the channel will be
		// withdrawn already and we're done.
		log.Debug("Channel already withdrawn.")
		return nil
	}

	if err := c.machine.SetRegistered(ctx); err != nil {
		return errors.WithMessage(err, "setting machine to Registered phase")
	}

	return c.settle(ctx, false)
}

// Settle settles the channel: it is made sure that the current state is
// registered and the final balance withdrawn. This call blocks until the
// channel has been successfully withdrawn.
func (c *Channel) Settle(ctx context.Context) error {
	return c.settleLocked(ctx, false)
}

// SettleSecondary settles the channel: it is made sure that the current state
// is registered and the final balance withdrawn. This call blocks until the
// channel has been successfully withdrawn.
//
// SettleSecondary is a variant of Settle that can be called when
// collaboratively settling a channel at the same time as the other channel
// peers. The initiator of the channel settlement should call Settle, whereas
// all responders can call SettleSecondary. The blockchain backend might then
// run an optimized settlement protocol that possibly saves sending unnecessary
// duplicate transactions in parallel. If the initiator is maliciously not
// sending the required transactions, the backend guarantees that it will
// eventually send the required transactions, so it is always safe to call
// SettleSecondary.
func (c *Channel) SettleSecondary(ctx context.Context) error {
	return c.settleLocked(ctx, true)
}

// settleLocked calls settle with the channel mutex locked and the context also
// set to cancel when the client is closed.
func (c *Channel) settleLocked(ctx context.Context, secondary bool) error {
	if !c.machMtx.TryLockCtx(ctx) {
		return errors.Errorf("locking machine mutex in time: %v", ctx.Err())
	}
	defer c.machMtx.Unlock()
	// Wrap the context to make sure that the settle call stops as soon as the
	// channel controller is closed.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c.OnClose(cancel)
	return c.settle(ctx, secondary)
}

// settle makes sure that the current state is registered and the final balance
// withdrawn. This call blocks until the channel has been successfully
// withdrawn.
//
// If the secondary flag is true and the channel is in a final state, the
// blockchain backend might run an optimized withdrawing protocol, possibly
// skipping the sending of unnecessary transactions, where the other user is
// assumed to be the initiator of the settlement process.
//
// The caller is expected to have locked the channel mutex.
func (c *Channel) settle(ctx context.Context, secondary bool) (err error) {
	if c.IsSubChannel() {
		err = c.subChannelSettleOptimistic(ctx)
	} else {
		err = c.ledgerChannelSettle(ctx, secondary)
	}
	if err != nil {
		return err
	}
	c.Log().Info("Withdrawal successful.")
	c.wallet.DecrementUsage(c.machine.Account().Address())
	return nil
}

func (c *Channel) ledgerChannelSettle(ctx context.Context, secondary bool) error {
	ver := c.machine.State().Version
	reg, err := func() (channel.AdjudicatorEvent, error) {
		ctx, cancel := context.WithTimeout(ctx, waitEventDuration)
		defer cancel()
		return c.registeredState(ctx)
	}()
	if err != nil {
		c.Log().Warnf("getting remote state: %v", err)
	}

	// If the machine is at least in phase Registered, reg shouldn't be nil. We
	// still catch this case to be future proof.
	if !c.machine.IsRegistered() || err != nil || reg.Version() < ver {
		if reg != nil && reg.Version() < ver {
			c.Log().Warnf("Lower version %d (< %d) registered, refuting...", reg.Version(), ver)
		}
		if err := c.register(ctx); err != nil {
			return errors.WithMessage(err, "registering")
		}
		c.Log().Info("Channel state registered.")
	}

	reg, err = c.registeredState(ctx)
	if err != nil {
		return errors.WithMessage(err, "getting remote state after registering")
	}

	if !reg.Timeout().IsElapsed(ctx) {
		if c.machine.State().IsFinal {
			c.Log().Warnf(
				"Unexpected withdrawal timeout while settling final state. Waiting until %v.",
				reg.Timeout)
		} else {
			c.Log().Infof("Waiting until %v for withdrawal.", reg.Timeout)
		}

		if err := reg.Timeout().Wait(ctx); err != nil {
			return errors.WithMessage(err, "waiting for timeout")
		}
	}

	return errors.WithMessage(c.withdraw(ctx, secondary), "withdrawing")
}

func (c *Channel) registeredState(ctx context.Context) (channel.AdjudicatorEvent, error) {
	sub, err := c.adjudicator.Subscribe(ctx, c.Params())
	if err != nil {
		return nil, errors.WithMessage(err, "creating event subscription")
	}
	defer sub.Close()

	ec := make(chan channel.AdjudicatorEvent)
	go func() {
		ec <- sub.Next()
	}()

	var e channel.AdjudicatorEvent
	select {
	case e = <-ec:
		sub.Close()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return e, sub.Err()
}

// register calls Register on the adjudicator with the current channel state and
// progresses the machine phases. When successful, the resulting RegisteredEvent
// is saved to the phase machine.
//
// The caller is expected to have locked the channel mutex.
func (c *Channel) register(ctx context.Context) error {
	if err := c.machine.SetRegistering(ctx); err != nil {
		return err
	}

	reg, err := c.adjudicator.Register(ctx, c.machine.AdjudicatorReq())
	if err != nil {
		return errors.WithMessage(err, "calling Register")
	}
	if ver := c.machine.State().Version; reg.Version() != ver {
		return errors.Errorf(
			"unexpected version %d registered, expected %d", reg.Version(), ver)
	}

	return c.machine.SetRegistered(ctx)
}

// withdraw calls Withdraw on the adjudicator with the current channel state and
// progresses the machine phases.
//
// The caller is expected to have locked the channel mutex.
func (c *Channel) withdraw(ctx context.Context, secondary bool) error {
	if err := c.machine.SetWithdrawing(ctx); err != nil {
		return err
	}

	req := c.machine.AdjudicatorReq()
	req.Secondary = secondary
	if err := c.adjudicator.Withdraw(ctx, req, nil); err != nil {
		return errors.WithMessage(err, "calling Withdraw")
	}

	return c.machine.SetWithdrawn(ctx)
}
