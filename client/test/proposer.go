// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/apps/payment"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
)

// Proposer is a test client role. He proposes the new channel.
type Proposer struct {
	role
}

// NewProposer creates a new party that executes the Proposer protocol.
func NewProposer(setup RoleSetup, t *testing.T, numStages int) *Proposer {
	return &Proposer{role: makeRole(setup, t, numStages)}
}

// Execute executes the Proposer protocol.
func (r *Proposer) Execute(cfg ExecConfig, exec func(ExecConfig, *paymentChannel)) {
	rng := rand.New(rand.NewSource(0x471CE))
	assert := assert.New(r.t)
	// We don't start the proposal listener because Proposer only sends proposals

	initBals := &channel.Allocation{
		Assets:   []channel.Asset{cfg.Asset},
		Balances: [][]channel.Bal{cfg.InitBals[:]},
	}
	prop := &client.ChannelProposal{
		ChallengeDuration: 60,           // 60 sec
		Nonce:             new(big.Int), // nonce 0
		ParticipantAddr:   r.setup.Wallet.NewRandomAccount(rng).Address(),
		AppDef:            payment.AppDef(),
		InitData:          new(payment.NoData),
		InitBals:          initBals,
		PeerAddrs:         cfg.PeerAddrs[:],
	}

	// send channel proposal
	ch, err := r.ProposeChannel(prop)
	assert.NoError(err)
	assert.NotNil(ch)
	if err != nil {
		return
	}
	r.log.Infof("New Channel opened: %v", ch.Channel)

	// start update handler
	handleDone := make(chan struct{})
	go func() {
		defer close(handleDone)
		r.log.Info("Starting request handler")
		// Note that we don't get any incoming proposal but still have to set a
		// handler.
		r.Handle(r.AcceptAllPropHandler(rng), r.UpdateHandler())
		r.log.Debug("Request handler returned.")
	}()
	defer func() {
		r.log.Debug("Waiting for request handler to return...")
		<-handleDone
	}()

	exec(cfg, ch)

	assert.NoError(ch.Close())
	assert.NoError(r.Close())
}
