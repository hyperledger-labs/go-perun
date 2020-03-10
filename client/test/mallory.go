// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/apps/payment"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	wallettest "perun.network/go-perun/wallet/test"
)

// Mallory is a test client role. She proposes the new channel.
type Mallory struct {
	Role
	rng *rand.Rand
}

// NewMallory creates a new party that executes the Mallory protocol.
func NewMallory(setup RoleSetup, t *testing.T) *Mallory {
	rng := rand.New(rand.NewSource(0x471CF))
	propHandler := newAcceptAllPropHandler(rng, setup.Timeout)
	role := &Mallory{
		Role: MakeRole(setup, propHandler, t, 3),
		rng:  rng,
	}

	propHandler.log = role.Role.log
	return role
}

// Execute executes the Mallory protocol.
func (r *Mallory) Execute(cfg ExecConfig) {
	assert := assert.New(r.t)
	we, _ := r.Idxs(cfg.PeerAddrs)
	// We don't start the proposal listener because Mallory only sends proposals

	initBals := &channel.Allocation{
		Assets:   []channel.Asset{cfg.Asset},
		Balances: [][]*big.Int{cfg.InitBals[:]},
	}
	prop := &client.ChannelProposal{
		ChallengeDuration: 60,           // 1 min
		Nonce:             new(big.Int), // nonce 0
		Account:           wallettest.NewRandomAccount(r.rng),
		AppDef:            payment.AppDef(),
		InitData:          new(payment.NoData),
		InitBals:          initBals,
		PeerAddrs:         cfg.PeerAddrs[:],
	}

	// send channel proposal
	_ch, err := func() (*client.Channel, error) {
		ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
		defer cancel()
		return r.ProposeChannel(ctx, prop)
	}()
	assert.NoError(err)
	assert.NotNil(_ch)
	if err != nil {
		return
	}
	ch := newPaymentChannel(_ch, &r.Role)
	r.log.Infof("New Channel opened: %v", ch.Channel)

	// AdjudicatorReq for version 0
	req0 := client.NewTestChannel(_ch).AdjudicatorReq()

	// 1st stage - channel controller set up
	r.waitStage()

	// Mallory sends some updates to Carol
	for i := 0; i < cfg.NumUpdates[we]; i++ {
		ch.sendTransfer(cfg.TxAmounts[we], fmt.Sprintf("Mallory#%d", i))
	}
	// 2nd stage - txs sent
	r.waitStage()

	// Register version 0 AdjudicatorReq
	challengeDuration := time.Duration(prop.ChallengeDuration) * time.Second
	regCtx, regCancel := context.WithTimeout(context.Background(), r.timeout)
	defer regCancel()
	r.log.Debug("Registering version 0 state.")
	reg0, err := r.setup.Adjudicator.Register(regCtx, req0)
	assert.NoError(err)
	assert.NotNil(reg0)
	r.log.Debugln("<Registered> ver 0: ", reg0)

	// within the challenge duration, Carol should refute.
	subCtx, subCancel := context.WithTimeout(context.Background(), r.timeout+challengeDuration)
	defer subCancel()
	sub, err := r.setup.Adjudicator.SubscribeRegistered(subCtx, ch.Params())

	// 3rd stage - wait until Carol has refuted
	r.waitStage()

	assert.True(reg0.Timeout.IsElapsed(),
		"Carol's refutation should already have progressed past the timeout.")
	reg := sub.Next() // should be event caused by Carol's refutation.
	assert.NoError(sub.Close())
	assert.NoError(sub.Err())
	assert.NotNil(reg)
	r.log.Debugln("<Registered> refuted: ", reg)
	if reg != nil {
		assert.Equal(ch.State().Version, reg.Version, "expected refutation with current version")
		waitCtx, waitCancel := context.WithTimeout(context.Background(), r.timeout+challengeDuration)
		defer waitCancel()
		// refutation increased the timeout.
		assert.NoError(reg.Timeout.Wait(waitCtx))
	}

	wdCtx, wdCancel := context.WithTimeout(context.Background(), r.timeout)
	defer wdCancel()
	err = r.setup.Adjudicator.Withdraw(wdCtx, req0)
	assert.Error(err, "withdrawing should fail because Carol should have refuted.")

	// settling current version should work
	ch.settleChan()

	assert.NoError(ch.Close())
	assert.NoError(r.Close())
}
