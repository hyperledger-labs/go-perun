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

// Package test contains helpers for testing the client
package test

import (
	"context"
	"errors"
	"io"
	"math/big"
	"math/rand"
	"sync"
	"testing"
	"time"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/client"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/watcher"
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
		timeout           time.Duration
		log               log.Logger
		t                 *testing.T
		numStages         int
		stages            Stages
		challengeDuration uint64
	}

	channelMap struct {
		entries map[channel.ID]*paymentChannel
		sync.RWMutex
	}

	// BalanceReader can be used to read state from a ledger.
	BalanceReader interface {
		Balance(p wallet.Address, a channel.Asset) channel.Bal
	}

	// RoleSetup contains the injectables for setting up the client.
	RoleSetup struct {
		Name              string
		Identity          wire.Account
		Bus               wire.Bus
		Funder            channel.Funder
		Adjudicator       channel.Adjudicator
		Watcher           watcher.Watcher
		Wallet            wallettest.Wallet
		PR                persistence.PersistRestorer // Optional PersistRestorer
		Timeout           time.Duration               // Timeout waiting for other role, not challenge duration
		BalanceReader     BalanceReader
		ChallengeDuration uint64
	}

	// ExecConfig contains additional config parameters for the tests.
	ExecConfig interface {
		Peers() [2]wire.Address   // must match the RoleSetup.Identity's
		Asset() channel.Asset     // single Asset to use in this channel
		InitBals() [2]*big.Int    // channel deposit of each role
		App() client.ProposalOpts // must be either WithApp or WithoutApp
	}

	// BaseExecConfig contains base config parameters.
	BaseExecConfig struct {
		peers    [2]wire.Address     // must match the RoleSetup.Identity's
		asset    channel.Asset       // single Asset to use in this channel
		initBals [2]*big.Int         // channel deposit of each role
		app      client.ProposalOpts // must be either WithApp or WithoutApp
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

const twoPartyTestTimeout = 10 * time.Second

// ExecuteTwoPartyTest executes the specified client test.
func ExecuteTwoPartyTest(t *testing.T, role [2]Executer, cfg ExecConfig) {
	log.Info("Starting two-party test")

	// enable stages synchronization
	stages := role[0].EnableStages()
	role[1].SetStages(stages)

	numClients := len(role)
	done := make(chan struct{}, numClients)

	// start clients
	for i := 0; i < numClients; i++ {
		go func(i int) {
			log.Infof("Executing role %d", i)
			role[i].Execute(cfg)
			done <- struct{}{} // signal client done
		}(i)
	}

	// wait for clients to finish or timeout
	timeout := time.After(twoPartyTestTimeout)
	for clientsRunning := numClients; clientsRunning > 0; clientsRunning-- {
		select {
		case <-done: // wait for client done signal
		case <-timeout:
			t.Fatal("twoPartyTest timeout")
		}
	}

	log.Info("Two-party test done")
}

// MakeBaseExecConfig creates a new BaseExecConfig.
func MakeBaseExecConfig(
	peers [2]wire.Address,
	asset channel.Asset,
	initBals [2]*big.Int,
	app client.ProposalOpts,
) BaseExecConfig {
	return BaseExecConfig{
		peers:    peers,
		asset:    asset,
		initBals: initBals,
		app:      app,
	}
}

// Peers returns the peer addresses.
func (c *BaseExecConfig) Peers() [2]wire.Address {
	return c.peers
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
		chans:             &channelMap{entries: make(map[channel.ID]*paymentChannel)},
		setup:             setup,
		timeout:           setup.Timeout,
		t:                 t,
		numStages:         numStages,
		challengeDuration: setup.ChallengeDuration,
	}
	cl, err := client.New(r.setup.Identity.Address(),
		r.setup.Bus, r.setup.Funder, r.setup.Adjudicator, r.setup.Wallet, r.setup.Watcher)
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
		r.log.Panic("number of stages don't match")
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
	acceptNextPropHandler struct {
		r     *role
		props chan proposalAndResponder
		rng   *rand.Rand
	}

	proposalAndResponder struct {
		prop client.ChannelProposal
		res  *client.ProposalResponder
	}
)

// GoHandle starts the handler routine on the current client and returns a
// wait() function with which it can be waited for the handler routine to stop.
func (r *role) GoHandle(rng *rand.Rand) (h *acceptNextPropHandler, wait func()) {
	done := make(chan struct{})
	propHandler := r.acceptNextPropHandler(rng)
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

func (r *role) LedgerChannelProposal(rng *rand.Rand, cfg ExecConfig) *client.LedgerChannelProposal {
	if !cfg.App().SetsApp() {
		r.log.Panic("Invalid ExecConfig: App does not specify an app.")
	}

	cfgInitBals := cfg.InitBals()
	initBals := &channel.Allocation{
		Assets:   []channel.Asset{cfg.Asset()},
		Balances: channel.Balances{cfgInitBals[:]},
	}
	cfgPeerAddrs := cfg.Peers()
	prop, err := client.NewLedgerChannelProposal(
		r.challengeDuration,
		r.setup.Wallet.NewRandomAccount(rng).Address(),
		initBals,
		cfgPeerAddrs[:],
		client.WithNonceFrom(rng),
		cfg.App())
	if err != nil {
		r.log.Panic("Error generating random channel proposal: " + err.Error())
	}
	return prop
}

func (r *role) SubChannelProposal(
	rng io.Reader,
	cfg ExecConfig,
	parent *client.Channel,
	initBals *channel.Allocation,
	app client.ProposalOpts,
) *client.SubChannelProposal {
	if !cfg.App().SetsApp() {
		r.log.Panic("Invalid ExecConfig: App does not specify an app.")
	}
	prop, err := client.NewSubChannelProposal(
		parent.ID(),
		r.challengeDuration,
		initBals,
		client.WithNonceFrom(rng),
		app,
	)
	if err != nil {
		r.log.Panic("Error generating random sub-channel proposal: " + err.Error())
	}
	return prop
}

func (r *role) acceptNextPropHandler(rng *rand.Rand) *acceptNextPropHandler {
	return &acceptNextPropHandler{
		r:     r,
		props: make(chan proposalAndResponder),
		rng:   rng,
	}
}

func (h *acceptNextPropHandler) HandleProposal(prop client.ChannelProposal, res *client.ProposalResponder) {
	select {
	case h.props <- proposalAndResponder{prop, res}:
	case <-time.After(h.r.setup.Timeout):
		h.r.t.Error("proposal response timeout") // Should be fatal, but cannot do this in go routine
	}
}

func (h *acceptNextPropHandler) Next() (*paymentChannel, error) {
	var prop client.ChannelProposal
	var res *client.ProposalResponder

	select {
	case pr := <-h.props:
		prop = pr.prop
		res = pr.res
	case <-time.After(h.r.setup.Timeout):
		return nil, errors.New("timeout passed")
	}

	h.r.log.Infof("Accepting incoming channel request: %v", prop)
	ctx, cancel := context.WithTimeout(context.Background(), h.r.setup.Timeout)
	defer cancel()

	var acc client.ChannelProposalAccept
	switch p := prop.(type) {
	case *client.LedgerChannelProposal:
		part := h.r.setup.Wallet.NewRandomAccount(h.rng).Address()
		acc = p.Accept(part, client.WithNonceFrom(h.rng))
		h.r.log.Debugf("Accepting ledger channel proposal with participant: %v", part)

	case *client.SubChannelProposal:
		acc = p.Accept(client.WithNonceFrom(h.rng))
		h.r.log.Debug("Accepting sub-channel proposal")

	default:
		panic("invalid proposal type")
	}

	ch, err := res.Accept(ctx, acc)
	if err != nil {
		return nil, err
	}
	// Client.OnNewChannel callback adds paymentChannel wrapper to the chans map
	payCh, ok := h.r.chans.get(ch.ID())
	if !ok {
		panic("channel not found")
	}
	return payCh, nil
}

type roleUpdateHandler role

func (r *role) UpdateHandler() *roleUpdateHandler { return (*roleUpdateHandler)(r) }

// HandleUpdate implements the Role as its own UpdateHandler.
func (h *roleUpdateHandler) HandleUpdate(_ *channel.State, up client.ChannelUpdate, res *client.UpdateResponder) {
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
