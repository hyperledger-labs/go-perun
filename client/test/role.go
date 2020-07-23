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

// Package test contains helpers for testing the client
package test

import (
	"context"
	"errors"
	"math/big"
	"math/rand"
	"sync"
	"testing"
	"time"

	"perun.network/go-perun/apps/payment"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/client"
	"perun.network/go-perun/log"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
)

type (
	// A Role is a client.Client together with a protocol execution path.
	role struct {
		*client.Client
		chans   map[channel.ID]*paymentChannel
		newChan func(*paymentChannel) // new channel callback
		setup   RoleSetup
		// we use the Client as Closer
		timeout   time.Duration
		log       log.Logger
		t         *testing.T
		numStages int
		stages    Stages
	}

	// RoleSetup contains the injectables for setting up the client.
	RoleSetup struct {
		Name        string
		Identity    wire.Account
		Bus         wire.Bus
		Funder      channel.Funder
		Adjudicator channel.Adjudicator
		Wallet      wallettest.Wallet
		PR          persistence.PersistRestorer // Optional PersistRestorer
		Timeout     time.Duration               // Timeout waiting for other role, not challenge duration
	}

	// ExecConfig contains additional config parameters for the tests.
	ExecConfig struct {
		PeerAddrs  [2]wire.Address // must match the RoleSetup.Identity's
		Asset      channel.Asset   // single Asset to use in this channel
		InitBals   [2]*big.Int     // channel deposit of each role
		NumUpdates [2]int          // how many updates each role sends
		TxAmounts  [2]*big.Int     // amounts that are to be sent by each role
	}

	// An Executer is a Role that can execute a protocol.
	Executer interface {
		// Execute executes the protocol according to the given configuration.
		Execute(cfg ExecConfig)
		// EnableStages enables role synchronization.
		EnableStages() Stages
		// SetStages enables role synchronization using the given stages.
		SetStages(Stages)
	}

	// Stages are used to synchronize multiple roles.
	Stages = []sync.WaitGroup
)

// makeRole creates a client for the given setup and wraps it into a Role.
func makeRole(setup RoleSetup, t *testing.T, numStages int) (r role) {
	r = role{
		chans:     make(map[channel.ID]*paymentChannel),
		setup:     setup,
		timeout:   setup.Timeout,
		t:         t,
		numStages: numStages,
	}
	cl, err := client.New(r.setup.Identity.Address(), r.setup.Bus, r.setup.Funder, r.setup.Adjudicator, r.setup.Wallet)
	if err != nil {
		t.Fatal("Error creating client: ", err)
	}
	r.setClient(cl) // init client
	return r
}

func (r *role) setClient(cl *client.Client) {
	if r.setup.PR != nil {
		cl.EnablePersistence(r.setup.PR)
	}
	cl.OnNewChannel(func(_ch *client.Channel) {
		ch := newPaymentChannel(_ch, r)
		r.chans[ch.ID()] = ch
		if r.newChan != nil {
			r.newChan(ch) // forward callback
		}
	})
	r.Client = cl
	// Append role field to client logger and set role logger to client logger.
	r.log = log.AppendField(cl, "role", r.setup.Name)
}

func (r *role) OnNewChannel(callback func(ch *paymentChannel)) {
	r.newChan = callback
}

// EnableStages optionally enables the synchronization of this role at different
// stages of the Execute protocol. EnableStages should be called on a single
// role and the resulting slice set on all remaining roles by calling SetStages
// on them.
func (r *role) EnableStages() Stages {
	r.stages = make(Stages, r.numStages)
	for i := range r.stages {
		r.stages[i].Add(1)
	}
	return r.stages
}

// SetStages optionally sets a slice of WaitGroup barriers to wait on at
// different stages of the Execute protocol. It should be created by any role by
// calling EnableStages().
func (r *role) SetStages(st Stages) {
	if len(st) != r.numStages {
		panic("number of stages don't match")
	}

	r.stages = st
	for i := range r.stages {
		r.stages[i].Add(1)
	}
}

func (r *role) waitStage() {
	if r.stages != nil {
		r.numStages--
		stage := &r.stages[r.numStages]
		stage.Done()
		stage.Wait()
	}
}

// Idxs maps the passed addresses to the indices in the 2-party-channel. If the
// setup's Identity is not found in peers, Idxs panics.
func (r *role) Idxs(peers [2]wire.Address) (our, their int) {
	if r.setup.Identity.Address().Equals(peers[0]) {
		return 0, 1
	} else if r.setup.Identity.Address().Equals(peers[1]) {
		return 1, 0
	}
	panic("identity not in peers")
}

// ProposeChannel sends the channel proposal req. It times out after the timeout
// specified in the Role's setup.
func (r *role) ProposeChannel(req *client.ChannelProposal) (*paymentChannel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	_ch, err := r.Client.ProposeChannel(ctx, req)
	if err != nil {
		return nil, err
	}
	// Client.OnNewChannel callback adds paymentChannel wrapper to the chans map
	return r.chans[_ch.ID()], nil
}

type (
	// acceptAllPropHandler is a channel proposal handler that accepts all channel
	// requests. It generates a random account for each channel.
	// Each accepted channel is put on the chans go channel.
	acceptAllPropHandler struct {
		r     *role
		chans chan channelAndError
		rng   *rand.Rand
	}

	// channelAndError bundles the return parameters of ProposalResponder.Accept
	// to be able to send them over a channel.
	channelAndError struct {
		channel *client.Channel
		err     error
	}
)

// GoHandle starts the handler routine on the current client and returns a
// wait() function with which it can be waited for the handler routine to stop.
func (r *role) GoHandle(rng *rand.Rand) (h *acceptAllPropHandler, wait func()) {
	done := make(chan struct{})
	propHandler := r.AcceptAllPropHandler(rng)
	go func() {
		defer close(done)
		r.log.Info("Starting request handler.")
		r.Handle(propHandler, r.UpdateHandler())
		r.log.Debug("Request handler returned.")
	}()

	return propHandler, func() {
		r.log.Debug("Waiting for request handler to return...")
		<-done
	}
}

func (r *role) ChannelProposal(rng *rand.Rand, cfg *ExecConfig) *client.ChannelProposal {
	initBals := &channel.Allocation{
		Assets:   []channel.Asset{cfg.Asset},
		Balances: [][]channel.Bal{cfg.InitBals[:]},
	}
	return &client.ChannelProposal{
		ChallengeDuration: 60, // 60 sec
		Nonce:             big.NewInt(rng.Int63()),
		ParticipantAddr:   r.setup.Wallet.NewRandomAccount(rng).Address(),
		AppDef:            payment.AppDef(),
		InitData:          new(payment.NoData),
		InitBals:          initBals,
		PeerAddrs:         cfg.PeerAddrs[:],
	}
}

// AcceptAllPropHandler returns a ProposalHandler that accepts all requests to
// this Role. The paymentChannel is saved to the Role's state upon acceptal. The
// rng is used to generate a new random account for accepting the proposal.
// Next can be called on the handler to wait for the next incoming proposal.
func (r *role) AcceptAllPropHandler(rng *rand.Rand) *acceptAllPropHandler {
	return &acceptAllPropHandler{
		r:     r,
		chans: make(chan channelAndError),
		rng:   rng,
	}
}

func (h *acceptAllPropHandler) HandleProposal(req *client.ChannelProposal, res *client.ProposalResponder) {
	h.r.log.Infof("Accepting incoming channel request: %v", req)
	ctx, cancel := context.WithTimeout(context.Background(), h.r.setup.Timeout)
	defer cancel()

	part := h.r.setup.Wallet.NewRandomAccount(h.rng).Address()
	h.r.log.Debugf("Accepting with participant: %v", part)
	ch, err := res.Accept(ctx, client.ProposalAcc{
		Participant: part,
	})
	h.chans <- channelAndError{ch, err}
}

// Next waits for the next incoming proposal. If the next proposal does not come
// in within one timeout period as specified in the Role's setup, it return nil
// and an error.
func (h *acceptAllPropHandler) Next() (*paymentChannel, error) {
	select {
	case ce := <-h.chans:
		if ce.err != nil {
			return nil, ce.err
		}
		// Client.OnNewChannel callback adds paymentChannel wrapper to the chans map
		return h.r.chans[ce.channel.ID()], nil
	case <-time.After(h.r.setup.Timeout):
		return nil, errors.New("timeout passed")
	}
}

type roleUpdateHandler role

func (r *role) UpdateHandler() *roleUpdateHandler { return (*roleUpdateHandler)(r) }

// HandleUpdate implements the Role as its own UpdateHandler
func (h *roleUpdateHandler) HandleUpdate(up client.ChannelUpdate, res *client.UpdateResponder) {
	ch, ok := h.chans[up.State.ID]
	if !ok {
		h.t.Errorf("unknown channel: %v", up.State.ID)
		ctx, cancel := context.WithTimeout(context.Background(), h.setup.Timeout)
		defer cancel()
		res.Reject(ctx, "unknown channel")
		return
	}

	ch.Handle(up, res)
}
