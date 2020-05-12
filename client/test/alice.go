// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test // import "perun.network/go-perun/client/test"

import (
	"fmt"
	"testing"
)

// Alice is a Proposer. She proposes the new channel.
type Alice struct {
	Proposer
}

// NewAlice creates a new Proposer that executes the Alice protocol.
func NewAlice(setup RoleSetup, t *testing.T) *Alice {
	return &Alice{Proposer: *NewProposer(setup, t, 4)}
}

// Execute executes the Alice protocol.
func (r *Alice) Execute(cfg ExecConfig) {
	r.Proposer.Execute(cfg, r.exec)
}

func (r *Alice) exec(cfg ExecConfig, ch *paymentChannel) {
	we, them := r.Idxs(cfg.PeerAddrs)
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
}
