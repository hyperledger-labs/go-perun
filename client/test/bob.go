// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test // import "perun.network/go-perun/client/test"

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Bob is a test client role. He accepts an incoming channel proposal.
type Bob struct {
	Role
	propHandler *acceptAllPropHandler
}

// NewBob creates a new party that executes the Bob protocol.
func NewBob(setup RoleSetup, t *testing.T) *Bob {
	rng := rand.New(rand.NewSource(0xB0B))
	propHandler := newAcceptAllPropHandler(rng, setup.Timeout)
	role := &Bob{
		Role:        MakeRole(setup, propHandler, t, 4),
		propHandler: propHandler,
	}

	propHandler.log = role.Role.log
	return role
}

// Execute executes the Bob protocol.
func (r *Bob) Execute(cfg ExecConfig) {
	assert := assert.New(r.t)
	we, them := r.Idxs(cfg.PeerAddrs)

	var listenWg sync.WaitGroup
	listenWg.Add(2)
	go func() {
		defer listenWg.Done()
		r.log.Info("Starting peer listener.")
		r.Listen(r.setup.Listener)
		r.log.Debug("Peer listener returned.")
	}()

	// receive one accepted proposal
	var chErr channelAndError
	select {
	case chErr = <-r.propHandler.chans:
	case <-time.After(r.timeout):
		r.t.Error("expected incoming channel proposal from Alice")
		return
	}
	assert.NoError(chErr.err)
	assert.NotNil(chErr.channel)
	if chErr.err != nil {
		return
	}
	ch := newPaymentChannel(chErr.channel, &r.Role)
	r.log.Info("New Channel opened: %v", ch.Channel)

	// start update handler
	go func() {
		defer listenWg.Done()
		r.log.Info("Starting update listener.")
		ch.ListenUpdates()
		r.log.Debug("Update listener returned.")
	}()
	defer func() {
		r.log.Debug("Waiting for listeners to return...")
		listenWg.Wait()
	}()
	// 1st stage - channel controller set up
	r.waitStage()

	// 1st Bob sends some updates to Alice
	for i := 0; i < cfg.NumUpdates[we]; i++ {
		ch.sendTransfer(cfg.TxAmounts[we], fmt.Sprintf("Bob#%d", i))
	}
	// 2nd stage
	r.waitStage()

	// 2nd Bob receives some updates from Alice
	for i := 0; i < cfg.NumUpdates[them]; i++ {
		ch.recvTransfer(cfg.TxAmounts[them], fmt.Sprintf("Alice#%d", i))
	}
	// 3rd stage
	r.waitStage()

	// 3rd Bob sends a final state
	ch.sendFinal()

	// 4th Settle channel
	ch.settleChan()

	// 4th final stage
	r.waitStage()

	assert.NoError(ch.Close())
	assert.NoError(r.Close())
}
