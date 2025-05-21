// Copyright 2020 - See NOTICE file for copyright holders.
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

package test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"perun.network/go-perun/client"
)

// MalloryCarolExecConfig contains config parameters for Mallory and Carol test.
type MalloryCarolExecConfig struct {
	BaseExecConfig
	NumPayments [2]int      // how many payments each role sends
	TxAmounts   [2]*big.Int // amounts that are to be sent/requested by each role
}

const malloryCarolNumStages = 3

// Mallory is a test client role. She proposes the new channel.
type Mallory struct {
	Proposer
}

// NewMallory creates a new party that executes the Mallory protocol.
func NewMallory(t *testing.T, setup RoleSetup) *Mallory {
	t.Helper()
	return &Mallory{Proposer: *NewProposer(t, setup, malloryCarolNumStages)}
}

// Execute executes the Mallory protocol.
func (r *Mallory) Execute(cfg ExecConfig) {
	r.Proposer.Execute(cfg, r.exec)
}

func (r *Mallory) exec(_cfg ExecConfig, ch *paymentChannel) {
	cfg := _cfg.(*MalloryCarolExecConfig)
	we, _ := r.Idxs(cfg.Peers())
	// AdjudicatorReq for version 0
	req0 := client.NewTestChannel(ch.Channel).AdjudicatorReq()

	// 1st stage - channel controller set up
	r.waitStage()

	// Mallory sends some updates to Carol
	for i := range cfg.NumPayments[we] {
		ch.sendTransfer(cfg.TxAmounts[we], fmt.Sprintf("Mallory#%d", i))
	}
	// 2nd stage - txs sent
	r.waitStage()

	// Register version 0 AdjudicatorReq
	duration := ch.Channel.Params().ChallengeDuration
	challengeDuration := time.Duration(duration) * time.Second //nolint:gosec
	regCtx, regCancel := context.WithTimeout(context.Background(), r.timeout)
	defer regCancel()
	r.log.Debug("Registering version 0 state.")
	r.RequireNoError(r.setup.Adjudicator.Register(regCtx, req0, nil))

	// within the challenge duration, Carol should refute.
	subCtx, subCancel := context.WithTimeout(context.Background(), r.timeout+challengeDuration)
	defer subCancel()
	sub, err := r.setup.Adjudicator.Subscribe(subCtx, ch.Params().ID())
	r.RequireNoError(err)

	// 3rd stage - wait until Carol has refuted
	r.waitStage()

	event := sub.Next() // should be event caused by Carol's refutation.
	r.RequireTrue(event != nil)
	r.RequireTruef(event.Timeout().IsElapsed(subCtx),
		"Carol's refutation should already have progressed past the timeout.")

	r.RequireNoError(sub.Close())
	r.RequireNoError(sub.Err())
	r.log.Debugln("<Registered> refuted: ", event)
	r.RequireTruef(ch.State().Version == event.Version(), "expected refutation with current version")
	waitCtx, waitCancel := context.WithTimeout(context.Background(), r.timeout+challengeDuration)
	defer waitCancel()
	// refutation increased the timeout.
	r.RequireNoError(event.Timeout().Wait(waitCtx))

	wdCtx, wdCancel := context.WithTimeout(context.Background(), r.timeout)
	defer wdCancel()
	err = r.setup.Adjudicator.Withdraw(wdCtx, req0, nil)
	r.RequireTruef(err != nil, "withdrawing should fail because Carol should have refuted")

	// settling current version should work
	ch.settle()
}
