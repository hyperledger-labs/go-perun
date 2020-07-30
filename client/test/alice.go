// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test // nolint: dupl

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
