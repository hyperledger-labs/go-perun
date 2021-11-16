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

// IsLedgerChannel returns whether the channel is a ledger channel.
func (c *Channel) IsLedgerChannel() bool {
	return c.Parent() == nil
}

// IsSubChannel returns whether the channel is a sub-channel.
func (c *Channel) IsSubChannel() bool {
	return c.Parent() != nil && c.equalParticipants(c.Parent())
}

func (c *Channel) equalParticipants(_c *Channel) bool {
	a, b := c.Peers(), _c.Peers()
	if len(a) != len(b) {
		return false
	}

	for i, _a := range a {
		if !_a.Equals(b[i]) {
			return false
		}
	}

	return true
}

func (c *Channel) fundSubChannel(ctx context.Context, id channel.ID, alloc *channel.Allocation) error {
	// We assume that the channel is locked.
	return c.updateBy(ctx, func(state *channel.State) error {
		// equal assets and sufficient balances are already checked when validating the sub-channel proposal

		// withdraw initial balances into sub-allocation
		state.Allocation.Balances = state.Allocation.Balances.Sub(alloc.Balances)
		state.AddSubAlloc(*channel.NewSubAlloc(id, alloc.Balances.Sum(), nil))

		return nil
	})
}

func (c *Channel) withdrawSubChannelIntoParent(ctx context.Context) error {
	if !c.IsSubChannel() {
		c.Log().Panic("not a sub-channel")
	} else if !c.machine.State().IsFinal {
		return errors.New("not final")
	}

	switch c.Idx() {
	case proposerIdx:
		err := c.Parent().withdrawSubChannel(ctx, c)
		return errors.WithMessage(err, "updating parent channel")
	case proposeeIdx:
		err := c.Parent().awaitSubChannelWithdrawal(ctx, c.ID())
		return errors.WithMessage(err, "awaiting parent channel update")
	default:
		c.Log().Panic("invalid participant index")
	}

	return nil
}

// withdrawSubChannel updates c so that the sub-channel allocation for
// subchannel is moved to c's balances. assumes that the subchannel is locked
// and finalized.
func (c *Channel) withdrawSubChannel(ctx context.Context, sub *Channel) error {
	// Lock machine while update is in progress.
	if !c.machMtx.TryLockCtx(ctx) {
		return errors.Errorf("locking machine mutex in time: %v", ctx.Err())
	}
	defer c.machMtx.Unlock()

	err := c.updateBy(ctx, func(parentState *channel.State) error {
		subAlloc, ok := parentState.SubAlloc(sub.ID())
		if !ok {
			c.Log().Panicf("sub-allocation %x not found", subAlloc.ID)
		}

		if !subAlloc.BalancesEqual(sub.machine.State().Allocation.Sum()) {
			c.Log().Panic("sub-allocation does not equal accumulated sub-channel outcome")
		}

		// asset types of this channel and parent channel are assumed to be the same
		for a, assetBalances := range sub.machine.State().Balances {
			for u, userBalance := range assetBalances {
				parentBalance := parentState.Allocation.Balances[a][u]
				parentBalance.Add(parentBalance, userBalance)
			}
		}

		if err := parentState.Allocation.RemoveSubAlloc(subAlloc); err != nil {
			c.Log().WithError(err).Panicf("removing sub-allocation with id %x", subAlloc.ID)
		}

		return nil
	})

	return errors.WithMessage(err, "update parent channel")
}

func (c *Channel) registerSubChannelFunding(id channel.ID, alloc []channel.Bal) {
	filter := func(cu ChannelUpdate) bool {
		expected := *channel.NewSubAlloc(id, alloc, nil)
		_, containedBefore := c.machine.State().SubAlloc(expected.ID)
		subAlloc, containedAfter := cu.State.SubAlloc(expected.ID)
		return !containedBefore && containedAfter && expected.Equal(&subAlloc) == nil
	}
	ui := newUpdateInterceptor(filter)
	c.subChannelFundings.Register(id, ui)
}

func (c *Channel) registerSubChannelSettlement(id channel.ID, bals [][]channel.Bal) {
	filter := func(cu ChannelUpdate) bool {
		_, containedBefore := c.machine.State().SubAlloc(id)
		_, containedAfter := cu.State.SubAlloc(id)
		equalBalances := c.machine.State().Balances.Add(bals).Equal(cu.State.Balances)

		return containedBefore && !containedAfter && equalBalances
	}
	ui := newUpdateInterceptor(filter)
	c.subChannelWithdrawals.Register(id, ui)
}

func (c *Channel) awaitSubChannelFunding(ctx context.Context, id channel.ID) error {
	return c.awaitSubChannelUpdate(ctx, id, c.subChannelFundings)
}

func (c *Channel) awaitSubChannelWithdrawal(ctx context.Context, id channel.ID) error {
	return c.awaitSubChannelUpdate(ctx, id, c.subChannelWithdrawals)
}

func (c *Channel) awaitSubChannelUpdate(ctx context.Context, id channel.ID, interceptors *updateInterceptors) error {
	ui, ok := interceptors.UpdateInterceptor(id)

	if !ok {
		return errors.New("not registered")
	}

	defer interceptors.Release(id)

	err := ui.Accept(ctx)
	return errors.WithMessage(err, "accepting update")
}
