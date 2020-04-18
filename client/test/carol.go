// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
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

// Carol is a test client role. He accepts an incoming channel proposal.
type Carol struct {
	Role
}

// NewCarol creates a new party that executes the Carol protocol.
func NewCarol(setup RoleSetup, t *testing.T) *Carol {
	return &Carol{Role: MakeRole(setup, t, 3)}
}

// Execute executes the Carol protocol.
func (r *Carol) Execute(cfg ExecConfig) {
	rng := rand.New(rand.NewSource(0xC407))
	assert := assert.New(r.t)
	_, them := r.Idxs(cfg.PeerAddrs)
	propHandler := newAcceptAllPropHandler(rng, r.setup.Timeout)

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
		r.t.Error("expected incoming channel proposal from Mallory")
		return
	}
	assert.NoError(chErr.err)
	assert.NotNil(chErr.channel)
	if chErr.err != nil {
		return
	}
	ch := newPaymentChannel(chErr.channel, &r.Role)
	r.log.Infof("New Channel opened: %v", ch.Channel)

	// start watcher
	watcher := make(chan error)
	go func() {
		r.log.Info("Starting channel watcher.")
		watcher <- ch.Watch()
		r.log.Debug("Channel watcher returned.")
	}()

	// start update handler
	go func() {
		defer listenWg.Done()
		r.log.Info("Starting update listener.")
		ch.ListenUpdates()
		r.log.Debug("Update listener returned.")
	}()
	// 1st stage - channel controller set up
	r.waitStage()

	// Carol receives some updates from Mallory
	for i := 0; i < cfg.NumUpdates[them]; i++ {
		ch.recvTransfer(cfg.TxAmounts[them], fmt.Sprintf("Mallory#%d", i))
	}
	// 2nd stage - txs received
	r.waitStage()

	r.log.Debug("Waiting for watcher to return...")
	select {
	case err := <-watcher:
		assert.NoError(err)
	case <-time.After(r.timeout):
		r.t.Error("expected watcher to return")
		return
	}

	// 3rd stage - channel settled
	r.waitStage()

	assert.NoError(ch.Close())
	assert.NoError(r.Close())
}
