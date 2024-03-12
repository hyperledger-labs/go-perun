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

package client

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire"
)

// withdrawVirtualChannel proposes to release the funds allocated to the
// specified virtual channel.
func (c *Channel) withdrawVirtualChannel(ctx context.Context, virtual *Channel) error {
	// Lock machine while update is in progress.
	if !c.machMtx.TryLockCtx(ctx) {
		return errors.Errorf("locking machine mutex in time: %v", ctx.Err())
	}
	defer c.machMtx.Unlock()

	state := c.state().Clone()
	state.Version++

	virtualAlloc, ok := state.SubAlloc(virtual.ID())
	if !ok {
		c.Log().Panicf("sub-allocation %x not found", virtualAlloc.ID)
	}

	if !virtualAlloc.BalancesEqual(virtual.state().Allocation.Sum()) {
		c.Log().Panic("sub-allocation does not equal accumulated sub-channel outcome")
	}

	virtualBalsRemapped := virtual.translateBalances(virtualAlloc.IndexMap)

	// We assume that the asset types of parent channel and virtual channel are the same.
	state.Allocation.Balances = state.Allocation.Balances.Add(virtualBalsRemapped)

	if err := state.Allocation.RemoveSubAlloc(virtualAlloc); err != nil {
		c.Log().WithError(err).Panicf("removing sub-allocation with id %x", virtualAlloc.ID)
	}

	err := c.updateGeneric(ctx, state, func(mcu *ChannelUpdateMsg) wire.Msg {
		return &VirtualChannelSettlementProposalMsg{
			ChannelUpdateMsg: *mcu,
			Final: channel.SignedState{
				Params: virtual.Params(),
				State:  virtual.state(),
				Sigs:   virtual.machine.CurrentTX().Sigs,
			},
		}
	})

	return errors.WithMessage(err, "update parent channel")
}

type proposalAndResponder struct {
	prop *VirtualChannelSettlementProposalMsg
	resp *UpdateResponder
}

func (c *Client) handleVirtualChannelSettlementProposal(
	parent *Channel,
	prop *VirtualChannelSettlementProposalMsg,
	responder *UpdateResponder,
) {
	err := c.validateVirtualChannelSettlementProposal(parent, prop)
	if err != nil {
		c.rejectProposal(responder, err.Error())
	}

	ctx, cancel := context.WithTimeout(c.Ctx(), virtualSettlementTimeout)
	defer cancel()

	err = c.settlementWatcher.Await(ctx, &proposalAndResponder{
		prop: prop,
		resp: responder,
	})
	if err != nil {
		c.rejectProposal(responder, err.Error())
	}
}

func (c *Client) validateVirtualChannelSettlementProposal(
	parent *Channel,
	prop *VirtualChannelSettlementProposalMsg,
) error {
	// Validate parameters.
	if prop.Final.Params.ID() != prop.Final.State.ID {
		return errors.New("invalid parameters")
	}

	// Validate signatures.
	for i, sig := range prop.Final.Sigs {
		ok, err := channel.Verify(
			prop.Final.Params.Parts[i],
			prop.Final.State,
			sig,
		)
		if err != nil {
			return err
		} else if !ok {
			return errors.New("invalid signature")
		}
	}

	// Validate allocation.

	// Assert equal assets.
	if err := channel.AssertAssetsEqual(parent.state().Assets, prop.Final.State.Assets); err != nil {
		return errors.WithMessage(err, "assets do not match")
	}

	// Assert contained before and matching funds
	subAlloc, containedBefore := parent.state().SubAlloc(prop.Final.Params.ID())
	if !containedBefore || !subAlloc.BalancesEqual(prop.Final.State.Sum()) {
		return errors.New("virtual channel not allocated")
	}

	// Assert not contained after
	_, containedAfter := prop.State.SubAlloc(prop.Final.Params.ID())
	if containedAfter {
		return errors.New("virtual channel must not be de-allocated after update")
	}

	// Assert correct balances
	virtual := transformBalances(prop.Final.State.Balances, parent.state().NumParts(), subAlloc.IndexMap)
	correctBalances := parent.state().Balances.Add(virtual).Equal(prop.State.Balances)
	if !correctBalances {
		return errors.New("invalid balances")
	}

	return nil
}

func (c *Client) matchSettlementProposal(ctx context.Context, a, b interface{}) bool {
	var err error
	defer func() {
		if err != nil {
			c.log.Debug("matching settlement proposal:", err)
		}
	}()

	// Cast.
	inputs := []interface{}{a, b}
	props := make([]*VirtualChannelSettlementProposalMsg, len(inputs))
	resps := make([]*UpdateResponder, len(inputs))
	for i, x := range inputs {
		pr, ok := x.(*proposalAndResponder)
		if !ok {
			err = errors.Errorf("casting %d", i)
			return false
		}
		props[i], resps[i] = pr.prop, pr.resp
	}

	prop0 := props[0]

	for i, prop := range props {
		// Check final state.
		err = prop.Final.State.Equal(prop0.Final.State)
		if err != nil {
			err = errors.Errorf("checking state equality %d: %v", i, err)
			return false
		}
	}

	// Store settlement state and signature.
	virtual, err := c.Channel(prop0.Final.State.ID)
	if err != nil {
		return false
	}
	err = virtual.forceFinalState(ctx, prop0.Final)
	if err != nil {
		return false
	}

	// Accept parent update proposals and thereby persist parent channel state
	// before deleting the virtual channel from persistence.
	for i, resp := range resps {
		if err = resp.Accept(ctx); err != nil {
			err = errors.Errorf("accepting update %d: %v", i, err)
			return false
		}
	}

	// Close channel and remove from persistence.
	err = virtual.Close()
	if err != nil {
		return false
	}
	c.channels.Delete(virtual.ID())
	err = virtual.machine.SetWithdrawn(ctx)
	return true
}

func (c *Channel) forceFinalState(ctx context.Context, final channel.SignedState) error {
	if err := c.machine.ForceUpdate(ctx, final.State, hubIndex); err != nil {
		return err
	}
	for i, sig := range final.Sigs {
		if err := c.machine.AddSig(ctx, channel.Index(i), sig); err != nil {
			return err
		}
	}
	return c.machine.EnableFinal(ctx)
}
