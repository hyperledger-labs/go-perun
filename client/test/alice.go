// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test // import "perun.network/go-perun/client/test"

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/apps/payment"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	wallettest "perun.network/go-perun/wallet/test"
)

type Alice struct {
	Role
	rng *rand.Rand
}

func NewAlice(setup RoleSetup, t *testing.T) *Alice {
	rng := rand.New(rand.NewSource(0x471CE))
	propHandler := newAcceptAllPropHandler(rng, setup.Timeout)
	role := &Alice{
		Role: MakeRole(setup, propHandler, t, 4),
		rng:  rng,
	}

	propHandler.log = role.Role.log
	return role
}

func (r *Alice) Execute(cfg ExecConfig) {
	assert := assert.New(r.t)
	// We don't start the proposal listener because Alice only receives proposals

	initBals := &channel.Allocation{
		Assets: []channel.Asset{cfg.Asset},
		OfParts: [][]*big.Int{
			[]*big.Int{cfg.InitBals[0]}, // Alice
			[]*big.Int{cfg.InitBals[1]}, // Bob
		},
	}
	prop := &client.ChannelProposal{
		ChallengeDuration: 10,           // 10 sec
		Nonce:             new(big.Int), // nonce 0
		Account:           wallettest.NewRandomAccount(r.rng),
		AppDef:            payment.AppDef(),
		InitData:          new(payment.NoData),
		InitBals:          initBals,
		PeerAddrs:         cfg.PeerAddrs,
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
	r.log.Info("New Channel opened: %v", ch.Channel)

	// start update handler
	listenUpDone := make(chan struct{})
	go func() {
		defer close(listenUpDone)
		r.log.Info("Starting update listener")
		ch.ListenUpdates()
		r.log.Debug("Update listener returned.")
	}()
	defer func() {
		r.log.Debug("Waiting for update listener to return...")
		<-listenUpDone
	}()
	// 1st stage - channel controller set up
	r.waitStage()

	// 1st Alice receives some updates from Bob
	for i := 0; i < cfg.NumUpdatesBob; i++ {
		ch.recvTransfer(cfg.TxAmountBob, fmt.Sprintf("Bob#%d", i))
	}
	// 2nd stage
	r.waitStage()

	// 2nd Alice sends some updates to Bob
	for i := 0; i < cfg.NumUpdatesAlice; i++ {
		ch.sendTransfer(cfg.TxAmountAlice, fmt.Sprintf("Alice#%d", i))
	}
	// 3rd stage
	r.waitStage()

	// 3rd Alice receives final state from Bob
	ch.recvFinal()

	// 4th Settle channel
	ch.settleChan()

	// 4th final stage
	r.waitStage()

	assert.NoError(ch.Close())
	assert.NoError(r.Close())
}
