// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test // import "perun.network/go-perun/client/test"

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Carol is a Responder. She accepts an incoming channel proposal.
type Carol struct {
	Responder
}

// NewCarol creates a new Responder that executes the Carol protocol.
func NewCarol(setup RoleSetup, t *testing.T) *Carol {
	return &Carol{Responder: *NewResponder(setup, t, 3)}
}

// Execute executes the Carol protocol.
func (r *Carol) Execute(cfg ExecConfig) {
	r.Responder.Execute(cfg, r.exec)
}

func (r *Carol) exec(cfg ExecConfig, ch *paymentChannel) {
	assert := assert.New(r.t)
	_, them := r.Idxs(cfg.PeerAddrs)

	// start watcher
	watcher := make(chan error)
	go func() {
		r.log.Info("Starting channel watcher.")
		watcher <- ch.Watch()
		r.log.Debug("Channel watcher returned.")
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
}
