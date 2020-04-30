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
)

// Alice is a test client role. She proposes the new channel.
type Alice struct {
	Role
}

// NewAlice creates a new party that executes the Alice protocol.
func NewAlice(setup RoleSetup, t *testing.T) *Alice {
	return &Alice{Role: MakeRole(setup, t, 4)}
}

// Execute executes the Alice protocol.
func (r *Alice) Execute(cfg ExecConfig) {
	rng := rand.New(rand.NewSource(0x471CE))
	assert := assert.New(r.t)
	we, them := r.Idxs(cfg.PeerAddrs)
	// We don't start the proposal listener because Alice only sends proposals

	initBals := &channel.Allocation{
		Assets:   []channel.Asset{cfg.Asset},
		Balances: [][]channel.Bal{cfg.InitBals[:]},
	}
	prop := &client.ChannelProposal{
		ChallengeDuration: 60,           // 1 min
		Nonce:             new(big.Int), // nonce 0
		ParticipantAddr:   r.setup.Wallet.NewRandomAccount(rng).Address(),
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
	for i := 0; i < cfg.NumUpdates[them]; i++ {
		ch.recvTransfer(cfg.TxAmounts[them], fmt.Sprintf("Bob#%d", i))
	}
	// 2nd stage
	r.waitStage()

	// 2nd Alice sends some updates to Bob
	for i := 0; i < cfg.NumUpdates[we]; i++ {
		ch.sendTransfer(cfg.TxAmounts[we], fmt.Sprintf("Alice#%d", i))
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
