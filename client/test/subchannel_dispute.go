// Copyright 2021 - See NOTICE file for copyright holders.
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
	"fmt"
	"math/big"
	"testing"
	"time"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
)

const nStagesDisputeSusieTime = 4

// DisputeSusieTimExecConfig contains config parameters for sub-channel dispute test.
type DisputeSusieTimExecConfig struct {
	BaseExecConfig
	SubChannelFunds [2]*big.Int // sub-channel funding amounts
	TxAmount        *big.Int    // transaction amount
}

// DisputeSusie is a Proposer. She proposes the new channel.
type DisputeSusie struct {
	Proposer
}

// NewDisputeSusie creates a new Proposer that executes the DisputeSusie protocol.
func NewDisputeSusie(t *testing.T, setup RoleSetup) *DisputeSusie {
	t.Helper()
	return &DisputeSusie{Proposer: *NewProposer(t, setup, nStagesDisputeSusieTime)}
}

// Execute executes the DisputeSusie protocol.
func (r *DisputeSusie) Execute(cfg ExecConfig) {
	r.Proposer.Execute(cfg, r.exec)
}

func (r *DisputeSusie) exec(_cfg ExecConfig, ledgerChannel *paymentChannel) {
	rng := r.NewRng()
	cfg := _cfg.(*DisputeSusieTimExecConfig)
	ctx := r.Ctx()

	// Stage 1 - Wait for channel controller setup.
	r.waitStage()

	// Stage 2 - Open sub-channel.
	subChannel := ledgerChannel.openSubChannel(rng, cfg, cfg.SubChannelFunds[:], client.WithoutApp())
	subReq0 := client.NewTestChannel(subChannel.Channel).AdjudicatorReq() // Store AdjudicatorReq for version 0
	r.waitStage()

	// Stage 3 - Update channels.
	update := func(ch *paymentChannel) {
		txAmount := cfg.TxAmount
		ch.sendTransfer(txAmount, fmt.Sprintf("send %v", txAmount))
	}

	update(ledgerChannel)
	update(subChannel)

	r.waitStage()

	// Stage 4 - Attack.

	// Register version 0 AdjudicatorReq.
	r.log.Debug("Registering version 0 state.")
	reqLedger := client.NewTestChannel(subChannel.Channel).AdjudicatorReq() // Current ledger state.
	subState0 := channel.SignedState{
		Params: subReq0.Params,
		State:  subReq0.Tx.State,
		Sigs:   subReq0.Tx.Sigs,
	}
	r.RequireNoError(r.setup.Adjudicator.Register(ctx, reqLedger, []channel.SignedState{subState0}))

	// Within the challenge duration, other party should refute.
	sub, err := r.setup.Adjudicator.Subscribe(ctx, subChannel.Params().ID())
	r.RequireNoError(err)

	// Wait until other party has refuted.
	for {
		event := sub.Next()
		r.RequireTrue(event != nil)
		r.RequireTruef(
			event.Timeout().IsElapsed(ctx),
			"Refutation should already have progressed past the timeout.",
		)
		if event.Version() > 0 {
			r.RequireNoError(sub.Close())
			r.RequireNoError(sub.Err())
			r.log.Debugln("<Registered> refuted: ", event)
			r.RequireTruef(subChannel.State().Version == event.Version(), "expected refutation with current version")
			r.RequireNoError(event.Timeout().Wait(ctx)) // Refutation increased the timeout.
			break
		}
	}

	r.log.Debug("Attempt withdrawing refuted state.")
	m := channel.MakeStateMap()
	m.Add(subState0.State)
	err = r.setup.Adjudicator.Withdraw(ctx, reqLedger, m)
	r.RequireTruef(err != nil, "withdraw should fail because other party should have refuted.")

	// Settling current version should work.
	r.log.Debug("Settle.")
	ledgerChannel.settle() // Settles ledger channel with sub-channels.

	r.log.Debug("Done.")
	r.waitStage()
}

// DisputeTim is a Responder. He accepts incoming channel proposals and updates.
type DisputeTim struct {
	Responder
	registered chan *channel.RegisteredEvent
	subCh      channel.ID
}

// time to wait until a parent channel watcher becomes active.
const channelWatcherWait = 100 * time.Millisecond

// HandleAdjudicatorEvent is the callback for adjudicator event handling.
func (r *DisputeTim) HandleAdjudicatorEvent(e channel.AdjudicatorEvent) {
	r.log.Infof("HandleAdjudicatorEvent: channelID = %x, version = %v, type = %T", e.ID(), e.Version(), e)
	if e, ok := e.(*channel.RegisteredEvent); ok && e.ID() == r.subCh {
		r.registered <- e
	}
}

// NewDisputeTim creates a new Responder that executes the DisputeTim protocol.
func NewDisputeTim(t *testing.T, setup RoleSetup) *DisputeTim {
	t.Helper()
	return &DisputeTim{
		Responder:  *NewResponder(t, setup, nStagesDisputeSusieTime),
		registered: make(chan *channel.RegisteredEvent),
	}
}

// Execute executes the DisputeTim protocol.
func (r *DisputeTim) Execute(cfg ExecConfig) {
	r.Responder.Execute(cfg, r.exec)
}

func (r *DisputeTim) exec(_cfg ExecConfig, ledgerChannel *paymentChannel, propHandler *acceptNextPropHandler) {
	cfg := _cfg.(*DisputeSusieTimExecConfig)

	// Stage 1 - Wait for channel controller setup.
	r.waitStage()

	// Stage 2 - Open sub-channel.
	subChannel := ledgerChannel.acceptSubchannel(propHandler, cfg.SubChannelFunds[:])
	r.subCh = subChannel.ID()
	// Start ledger channel watcher.
	go func() {
		r.log.Info("Starting ledger channel watcher.")
		err := ledgerChannel.Watch(r)
		r.log.Infof("Ledger channel watcher returned: %v", err)
	}()
	// Start sub-channel watcher.
	time.Sleep(channelWatcherWait) // Wait until parent channel watcher active.
	go func() {
		r.log.Info("Starting sub-channel watcher.")
		err := subChannel.Watch(r)
		r.log.Infof("Sub-channel watcher returned: %v", err)
	}()
	r.waitStage()

	// Stage 3 - Update channels.
	acceptUpdate := func(ch *paymentChannel) {
		txAmount := cfg.TxAmount
		ch.recvTransfer(txAmount, fmt.Sprintf("receive %v", txAmount))
	}

	acceptUpdate(ledgerChannel)
	acceptUpdate(subChannel)

	r.waitStage()

	// Stage 4 - Refutation.

	// Wait for refutation.
	r.log.Debug("Waiting for registered event from refutation.")
	e := func() channel.AdjudicatorEvent {
		for {
			select {
			case e := <-r.registered:
				if e.Version() == subChannel.State().Version {
					return e
				}
			case <-r.Ctx().Done():
				r.RequireNoError(r.Ctx().Err())
			}
		}
	}()
	r.log.Debug("Received refutation event, waiting until ready to conclude.")
	r.RequireNoError(e.Timeout().Wait(r.Ctx()))

	r.log.Debug("Settle")
	ledgerChannel.settle() // Settles ledger channel with sub-channels.

	r.log.Debug("Done.")
	r.waitStage()
}
