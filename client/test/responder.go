// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Responder is a test client role. He accepts an incoming channel proposal.
type Responder struct {
	Role
}

// NewResponder creates a new party that executes the Responder protocol.
func NewResponder(setup RoleSetup, t *testing.T, numStages int) *Responder {
	return &Responder{Role: MakeRole(setup, t, numStages)}
}

// Execute executes the Responder protocol.
func (r *Responder) Execute(cfg ExecConfig, exec func(ExecConfig, *paymentChannel)) {
	rng := rand.New(rand.NewSource(0xB0B))
	assert := assert.New(r.t)
	propHandler := newAcceptAllPropHandler(rng, r.setup.Timeout, r.setup.Wallet)

	var listenWg sync.WaitGroup
	listenWg.Add(3)
	defer func() {
		r.log.Debug("Waiting for listeners to return...")
		listenWg.Wait()
	}()
	go func() {
		defer listenWg.Done()
		r.log.Info("Starting peer listener.")
		r.Listen(r.setup.Listener)
		r.log.Debug("Peer listener returned.")
	}()
	go func() {
		defer listenWg.Done()
		r.log.Info("Starting proposal handler.")
		r.HandleChannelProposals(propHandler)
		r.log.Debug("Proposal handler returned.")
	}()

	// receive one accepted proposal
	var chErr channelAndError
	select {
	case chErr = <-propHandler.chans:
	case <-time.After(r.timeout):
		r.t.Error("expected incoming channel proposal from Proposer")
		return
	}
	assert.NoError(chErr.err)
	assert.NotNil(chErr.channel)
	if chErr.err != nil {
		return
	}

	ch := newPaymentChannel(chErr.channel, &r.Role)
	r.log.Infof("New Channel opened: %v", ch.Channel)
	// start update handler
	go func() {
		defer listenWg.Done()
		r.log.Info("Starting update listener.")
		ch.ListenUpdates()
		r.log.Debug("Update listener returned.")
	}()

	exec(cfg, ch)

	assert.NoError(ch.Close())
	assert.NoError(r.Close())
}
