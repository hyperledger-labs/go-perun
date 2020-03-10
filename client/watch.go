// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
)

// Watch starts the channel watcher routine. It subscribes to RegisteredEvents
// on the adjudicator. If an outside event
func (c *Channel) Watch() {
	log := c.log.WithField("proc", "watcher")
	defer log.Info("Watcher returned.")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.OnClose(cancel)
	sub, err := c.adjudicator.SubscribeRegistered(ctx, c.Params())
	if err != nil {
		log.Errorf("failed to setup subscription to RegisteredEvents: %v", err)
		return
	}

	// Wait for on-chain event
	reg := sub.Next()
	log.Infof("New RegisteredEvent: %v", reg)
	if reg == nil {
		if err := sub.Err(); err != nil {
			log.Warnf("Subscription closed: %v", err)
		} else {
			log.Debug("Subscription closed without error.", err)
		}
		return
	}

	if err := c.handleRegisteredEvent(ctx, reg); err != nil {
		log.Errorf("Handling Registered failed: %v", err)
	}
}

// handleRegisteredEvent stores the passed RegisteredEvent to the machine and
// settles the channel.
func (c *Channel) handleRegisteredEvent(ctx context.Context, reg *channel.RegisteredEvent) error {
	log := c.log.WithField("proc", "watcher")
	c.machMtx.Lock() // lock machine while registering is in progress
	defer c.machMtx.Unlock()

	if c.machine.Phase() == channel.Withdrawn {
		// If a Settle call by the user caused this event, the channel will be
		// withdrawn already and we're done.
		log.Debug("Channel already withdrawn.")
		return nil
	}

	if err := c.machine.SetRegistered(reg); err != nil {
		return errors.WithMessage(err, "setting machine to Registered phase")
	}

	return c.settle(ctx)
}

// Settle settles the channel: it is made sure that the current state is
// registered and the final balance withdrawn. This call blocks until the
// channel has been successfully withdrawn.
func (c *Channel) Settle(ctx context.Context) error {
	c.machMtx.Lock()
	defer c.machMtx.Unlock()
	// Wrap the context to make sure that the settle call stops as soon as the
	// channel controller is closed.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c.OnClose(cancel)
	return c.settle(ctx)
}

// settle makes sure that the current state is registered and the final balance
// withdrawn. This call blocks until the channel has been successfully
// withdrawn.
//
// The caller is expected to have locked the channel mutex.
func (c *Channel) settle(ctx context.Context) error {
	ver, reg := c.machine.State().Version, c.machine.Registered()
	// If the machine is at least in phase Registered, reg shouldn't be nil. We
	// still catch this case to be future proof.
	if c.machine.Phase() < channel.Registered || reg == nil || reg.Version < ver {
		if reg != nil && reg.Version < ver {
			c.log.Warnf("Lower version %d (< %d) registered, refuting...", reg.Version, ver)
		}
		if err := c.register(ctx); err != nil {
			return errors.WithMessage(err, "registering")
		}
		c.log.Info("Channel state registered.")
	}

	reg = c.machine.Registered() // reload in case we called Register
	if reg.Timeout.After(time.Now()) {
		if c.machine.State().IsFinal {
			c.log.Warnf(
				"Unexpected withdrawal timeout while settling final state. Waiting until %v.",
				reg.Timeout)
		} else {
			c.log.Infof("Waiting until %v for withdrawal.", reg.Timeout)
		}

		timeout := time.After(time.Until(reg.Timeout))
		select {
		case <-timeout: // proceed normally with withdrawal
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "ctx done")
		}
	}

	if err := c.withdraw(ctx); err != nil {
		return errors.WithMessage(err, "withdrawing")
	}
	c.log.Info("Withdrawal successful.")
	return nil
}

// register calls Regsiter on the adjudicator with the current channel state and
// progresses the machine phases. When successful, the resulting RegisteredEvent
// is saved to the phase machine.
//
// The caller is expected to have locked the channel mutex.
func (c *Channel) register(ctx context.Context) error {
	if err := c.machine.SetRegistering(); err != nil {
		return err
	}

	reg, err := c.adjudicator.Register(ctx, c.machine.AdjudicatorReq())
	if err != nil {
		return errors.WithMessage(err, "calling Register")
	}
	if ver := c.machine.State().Version; reg.Version != ver {
		return errors.Errorf(
			"unexpected version %d registered, expected %d", reg.Version, ver)
	}

	return c.machine.SetRegistered(reg)
}

// withdraw calls Withdraw on the adjudicator with the current channel state and
// progresses the machine phases.
//
// The caller is expected to have locked the channel mutex.
func (c *Channel) withdraw(ctx context.Context) error {
	if err := c.machine.SetWithdrawing(); err != nil {
		return err
	}

	req := c.machine.AdjudicatorReq()
	if err := c.adjudicator.Withdraw(ctx, req); err != nil {
		return errors.WithMessage(err, "calling Withdraw")
	}

	return c.machine.SetWithdrawn()
}
