// Copyright 2025 - See NOTICE file for copyright holders.
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

	"perun.network/go-perun/client"
)

// SusieTimExecConfig contains config parameters for Susie and Tim test.
type SusieTimExecConfig struct {
	BaseExecConfig

	SubChannelFunds    [][2]*big.Int       // sub-channel funding amounts, also determines number of sub-channels, must be at least 1
	SubSubChannelFunds [][2]*big.Int       // sub-sub-channel funding amounts, also determines number of sub-sub-channels
	LeafChannelApp     client.ProposalOpts // app used in the leaf channels
	TxAmount           *big.Int            // transaction amount
}

// NewSusieTimExecConfig creates a new object from the given parameters.
func NewSusieTimExecConfig(
	base BaseExecConfig,
	subChannelFunds [][2]*big.Int,
	subSubChannelFunds [][2]*big.Int,
	leafChannelApp client.ProposalOpts,
	txAmount *big.Int,
) *SusieTimExecConfig {
	return &SusieTimExecConfig{
		BaseExecConfig:     base,
		SubChannelFunds:    subChannelFunds,
		SubSubChannelFunds: subSubChannelFunds,
		LeafChannelApp:     leafChannelApp,
		TxAmount:           txAmount,
	}
}

// Susie is a Proposer. She proposes the new channel.
type Susie struct {
	Proposer
}

const susieTimNumStates = 7

// NewSusie creates a new Proposer that executes the Susie protocol.
func NewSusie(t *testing.T, setup RoleSetup) *Susie {
	t.Helper()
	return &Susie{Proposer: *NewProposer(t, setup, susieTimNumStates)}
}

// Execute executes the Susie protocol.
func (r *Susie) Execute(cfg ExecConfig) {
	r.Proposer.Execute(cfg, r.exec)
}

func (r *Susie) exec(_cfg ExecConfig, ledgerChannel *paymentChannel) {
	cfg := _cfg.(*SusieTimExecConfig)
	rng := r.NewRng()

	// stage 1 - channel controller set up
	r.waitStage()

	// stage 2 - open subchannels
	openSubChannel := func(parentChannel *paymentChannel, funds []*big.Int, app client.ProposalOpts) *paymentChannel {
		return parentChannel.openSubChannel(rng, cfg, funds, app, cfg.backend)
	}

	var subChannels []*paymentChannel
	for i := range len(cfg.SubChannelFunds) {
		c := openSubChannel(ledgerChannel, cfg.SubChannelFunds[i][:], cfg.App())
		subChannels = append(subChannels, c)
	}

	var subSubChannels []*paymentChannel
	for i := range len(cfg.SubSubChannelFunds) {
		c := openSubChannel(subChannels[0], cfg.SubSubChannelFunds[i][:], cfg.LeafChannelApp)
		subSubChannels = append(subSubChannels, c)
	}

	r.waitStage()

	// stage 3 - update channels

	update := func(ch *paymentChannel) {
		txAmount := cfg.TxAmount
		ch.sendTransfer(txAmount, fmt.Sprintf("send %v", txAmount))
	}

	update(ledgerChannel)
	for _, ch := range subChannels {
		update(ch)
	}
	for _, ch := range subSubChannels {
		update(ch)
	}

	r.waitStage()

	// stage 4 - finalize subchannels 2, ..., NumSubChannels

	finalizeAndSettle := func(ch *paymentChannel) {
		ch.sendFinal()
		ch.settle()
	}

	for i, ch := range subChannels {
		if i == 0 {
			continue
		}
		finalizeAndSettle(ch)
	}

	r.waitStage()

	// stage 5 - finalize sub-channels 1.1, ..., 1.NumSubSubChannels

	for _, ch := range subSubChannels {
		finalizeAndSettle(ch)
	}

	r.waitStage()

	// stage 6 - finalize subchannel 1
	finalizeAndSettle(subChannels[0])

	r.waitStage()

	// stage 7 - finalize ledger channel

	finalizeAndSettle(ledgerChannel)

	r.waitStage()
}

// Tim is a Responder. He accepts incoming channel proposals and updates.
type Tim struct {
	Responder
}

// NewTim creates a new Responder that executes the Tim protocol.
func NewTim(t *testing.T, setup RoleSetup) *Tim {
	t.Helper()
	return &Tim{Responder: *NewResponder(t, setup, susieTimNumStates)}
}

// Execute executes the Tim protocol.
func (r *Tim) Execute(cfg ExecConfig) {
	r.Responder.Execute(cfg, r.exec)
}

func (r *Tim) exec(_cfg ExecConfig, ledgerChannel *paymentChannel, propHandler *acceptNextPropHandler) {
	cfg := _cfg.(*SusieTimExecConfig)

	// 1st stage - channel controller set up
	r.waitStage()

	// 2nd stage - open subchannels
	acceptNext := func(ch *paymentChannel, funds []*big.Int) *paymentChannel {
		return ch.acceptSubchannel(propHandler, funds)
	}

	var subChannels []*paymentChannel
	for i := range len(cfg.SubChannelFunds) {
		c := acceptNext(ledgerChannel, cfg.SubChannelFunds[i][:])
		subChannels = append(subChannels, c)
	}

	var subSubChannels []*paymentChannel
	for i := range len(cfg.SubSubChannelFunds) {
		c := acceptNext(subChannels[0], cfg.SubSubChannelFunds[i][:])
		subSubChannels = append(subSubChannels, c)
	}

	r.waitStage()

	// stage 3 - update channels

	acceptUpdate := func(ch *paymentChannel) {
		txAmount := cfg.TxAmount
		ch.recvTransfer(txAmount, fmt.Sprintf("receive %v", txAmount))
	}

	acceptUpdate(ledgerChannel)
	for _, ch := range subChannels {
		acceptUpdate(ch)
	}
	for _, ch := range subSubChannels {
		acceptUpdate(ch)
	}

	r.waitStage()

	// stage 4 - finalize sub-channels 2, ..., NumSubChannels

	finalizeAndSettle := func(ch *paymentChannel) {
		ch.recvFinal()
		ch.settleSecondary()
	}

	for i, ch := range subChannels {
		if i == 0 {
			continue
		}
		finalizeAndSettle(ch)
	}

	r.waitStage()

	// stage 5 - finalize sub-channels 1.1, ..., 1.NumSubSubChannels

	for _, ch := range subSubChannels {
		finalizeAndSettle(ch)
	}

	r.waitStage()

	// stage 6 - finalize subchannel 1
	finalizeAndSettle(subChannels[0])

	r.waitStage()

	// stage 7 - finalize ledger channel

	finalizeAndSettle(ledgerChannel)

	r.waitStage()
}
