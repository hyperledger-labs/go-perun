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
		chans   *channelMap
		newChan func(*paymentChannel) // new channel callback
		setup   RoleSetup
		// we use the Client as Closer
		timeout   time.Duration
		log       log.Logger
		t         *testing.T
		numStages int
		stages    Stages
	}

	channelMap struct {
		entries map[channel.ID]*paymentChannel
		sync.RWMutex
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
	ExecConfig interface {
		PeerAddrs() [2]wire.Address // must match the RoleSetup.Identity's
		Asset() channel.Asset       // single Asset to use in this channel
		InitBals() [2]*big.Int      // channel deposit of each role
		App() client.ProposalOpts   // must be either WithApp or WithoutApp
	}

	// BaseExecConfig contains base config parameters.
	BaseExecConfig struct {
		peerAddrs [2]wire.Address     // must match the RoleSetup.Identity's
		asset     channel.Asset       // single Asset to use in this channel
		initBals  [2]*big.Int         // channel deposit of each role
		app       client.ProposalOpts // must be either WithApp or WithoutApp
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

// MakeBaseExecConfig creates a new BaseExecConfig.
func MakeBaseExecConfig(
	peerAddrs [2]wire.Address,
	asset channel.Asset,
	initBals [2]*big.Int,
	app client.ProposalOpts,
) BaseExecConfig {
	return BaseExecConfig{
		peerAddrs: peerAddrs,
		asset:     asset,
		initBals:  initBals,
		app:       app,
	}
}

// PeerAddrs returns the peer addresses.
func (c *BaseExecConfig) PeerAddrs() [2]wire.Address {
	return c.peerAddrs
}

// Asset returns the asset.
func (c *BaseExecConfig) Asset() channel.Asset {
	return c.asset
}

// InitBals returns the initial balances.
func (c *BaseExecConfig) InitBals() [2]*big.Int {
	return c.initBals
}

// App returns the app.
func (c *BaseExecConfig) App() client.ProposalOpts {
	return c.app
}

// makeRole creates a client for the given setup and wraps it into a Role.
func makeRole(setup RoleSetup, t *testing.T, numStages int) (r role) {
	r = role{
		chans:     &channelMap{entries: make(map[channel.ID]*paymentChannel)},
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
		r.chans.add(ch)
		if r.newChan != nil {
			r.newChan(ch) // forward callback
		}
	})
	r.Client = cl
	// Append role field to client logger and set role logger to client logger.
	r.log = log.AppendField(cl, "role", r.setup.Name)
}

func (chs *channelMap) get(ch channel.ID) (_ch *paymentChannel, ok bool) {
	chs.RLock()
	defer chs.RUnlock()
	_ch, ok = chs.entries[ch]
	return
}

func (chs *channelMap) add(ch *paymentChannel) {
	chs.Lock()
	defer chs.Unlock()
	chs.entries[ch.ID()] = ch
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
func (r *role) Idxs(peers [2]wire.Address) (our, their channel.Index) {
	if r.setup.Identity.Address().Equals(peers[0]) {
		return 0, 1
	} else if r.setup.Identity.Address().Equals(peers[1]) {
		return 1, 0
	}
	panic("identity not in peers")
}

// ProposeChannel sends the channel proposal req. It times out after the timeout
// specified in the Role's setup.
func (r *role) ProposeChannel(req client.ChannelProposal) (*paymentChannel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	_ch, err := r.Client.ProposeChannel(ctx, req)
	if err != nil {
		return nil, err
	}
	// Client.OnNewChannel callback adds paymentChannel wrapper to the chans map
	ch, ok := r.chans.get(_ch.ID())
	if !ok {
		return ch, errors.New("channel not found")
	}
	return ch, nil
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

const challengeDuration = 60

func (r *role) LedgerChannelProposal(rng *rand.Rand, cfg ExecConfig) client.ChannelProposal {
	if !cfg.App().SetsApp() {
		r.log.Panic("Invalid ExecConfig: App does not specify an app.")
	}

	cfgInitBals := cfg.InitBals()
	initBals := &channel.Allocation{
		Assets:   []channel.Asset{cfg.Asset()},
		Balances: channel.Balances{cfgInitBals[:]},
	}
	cfgPeerAddrs := cfg.PeerAddrs()
	return client.NewLedgerChannelProposal(
		challengeDuration,
		r.setup.Wallet.NewRandomAccount(rng).Address(),
		initBals,
		cfgPeerAddrs[:],
		client.WithNonceFrom(rng),
		cfg.App())
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

func (h *acceptAllPropHandler) HandleProposal(req client.ChannelProposal, res *client.ProposalResponder) {
	h.r.log.Infof("Accepting incoming channel request: %v", req)
	ctx, cancel := context.WithTimeout(context.Background(), h.r.setup.Timeout)
	defer cancel()

	part := h.r.setup.Wallet.NewRandomAccount(h.rng).Address()
	h.r.log.Debugf("Accepting with participant: %v", part)
	acc := req.Base().NewChannelProposalAcc(part, client.WithNonceFrom(h.rng))
	ch, err := res.Accept(ctx, acc)
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
		ch, ok := h.r.chans.get(ce.channel.ID())
		if !ok {
			return ch, errors.New("channel not found")
		}
		return ch, nil
	case <-time.After(h.r.setup.Timeout):
		return nil, errors.New("timeout passed")
	}
}

type roleUpdateHandler role

func (r *role) UpdateHandler() *roleUpdateHandler { return (*roleUpdateHandler)(r) }

// HandleUpdate implements the Role as its own UpdateHandler.
func (h *roleUpdateHandler) HandleUpdate(up client.ChannelUpdate, res *client.UpdateResponder) {
	ch, ok := h.chans.get(up.State.ID)
	if !ok {
		h.t.Errorf("unknown channel: %v", up.State.ID)
		ctx, cancel := context.WithTimeout(context.Background(), h.setup.Timeout)
		defer cancel()
		res.Reject(ctx, "unknown channel")
		return
	}

	ch.Handle(up, res)
}
