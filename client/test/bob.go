// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"fmt"
	"testing"
)

// Bob is a Responder. He accepts an incoming channel proposal.
type Bob struct {
	Responder
}

// NewBob creates a new Responder that executes the Bob protocol.
func NewBob(setup RoleSetup, t *testing.T) *Bob {
	return &Bob{Responder: *NewResponder(setup, t, 4)}
}

// Execute executes the Bob protocol.
func (r *Bob) Execute(cfg ExecConfig) {
	r.Responder.Execute(cfg, r.exec)
}

func (r *Bob) exec(cfg ExecConfig, ch *paymentChannel) {
	we, them := r.Idxs(cfg.PeerAddrs)

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
}
