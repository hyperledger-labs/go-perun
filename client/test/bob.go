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

type Bob struct {
	Role
	propHandler *acceptAllPropHandler
}

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

func (r *Bob) Execute(cfg ExecConfig) {
	assert := assert.New(r.t)

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
		r.t.Fatal("expected incoming channel proposal from Alice")
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
	for i := 0; i < cfg.NumUpdatesBob; i++ {
		ch.sendTransfer(cfg.TxAmountBob, fmt.Sprintf("Bob#%d", i))
	}
	// 2nd stage
	r.waitStage()

	// 2nd Bob receives some updates from Alice
	for i := 0; i < cfg.NumUpdatesAlice; i++ {
		ch.recvTransfer(cfg.TxAmountAlice, fmt.Sprintf("Alice#%d", i))
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
