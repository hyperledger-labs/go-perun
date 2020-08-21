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

// Bob is a Responder. He accepts an incoming channel proposal.
type Bob struct {
	Responder
}

// NewBob creates a new Responder that executes the Bob protocol.
func NewBob(setup RoleSetup, t *testing.T) *Bob {
	return &Bob{Responder: *NewResponder(setup, t, 6)}
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
	for i := 0; i < cfg.NumPayments[we]; i++ {
		ch.sendTransfer(cfg.TxAmounts[we], we, fmt.Sprintf("Bob#%d", i))
	}

	// 2nd stage
	r.waitStage()

	// 2nd Bob receives some updates from Alice
	for i := 0; i < cfg.NumPayments[them]; i++ {
		ch.recvTransfer(cfg.TxAmounts[them], we, fmt.Sprintf("Alice#%d", i))
	}

	// 3rd stage
	r.waitStage()

	// 3rd Bob sends payment requests to Alice.
	for i := 0; i < cfg.NumRequests[we]; i++ {
		ch.sendTransfer(cfg.TxAmounts[we], them, fmt.Sprintf("Bob-req#%d", i))
	}

	// 4th stage
	r.waitStage()

	// 4th Bob receives for payment requests from Alice.
	for i := 0; i < cfg.NumRequests[them]; i++ {
		ch.recvTransfer(cfg.TxAmounts[them], them, fmt.Sprintf("Alice-req#%d", i))
	}

	// 5th stage
	r.waitStage()

	// 5rd Bob sends a final state
	ch.sendFinal()

	// 5th Settle channel
	ch.settleSecondary()

	// 6th final stage
	r.waitStage()
}
