// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test // import "perun.network/go-perun/client/test"

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/pkg/test"
)

// A ConnHub can be used to create related Listeners and Dialers.
// This is (almost) fulfilled by the peer/test.ConnHub but in order to be able
// to use possible other netorking infrastructure for the tests in the future,
// it is consumed as an interface in the client persistence tests.
type ConnHub interface {
	NewListener(addr peer.Address) peer.Listener
	NewDialer() peer.Dialer
}

type (
	multiClientRole struct {
		role
		hub ConnHub
	}

	// Petra is the Proposer in a Persistence test.
	Petra struct{ multiClientRole }

	// Robert is the Responder in a persistence test.
	Robert struct{ multiClientRole }
)

// ReplaceClient replaces the client instance of the Role. Useful for
// persistence testing.
func (r *multiClientRole) ReplaceClient() {
	dialer := r.hub.NewDialer()
	cl := client.New(r.setup.Identity, dialer, r.setup.Funder, r.setup.Adjudicator, r.setup.Wallet)
	r.setClient(cl)
}

func (r *multiClientRole) NewListener() peer.Listener {
	return r.hub.NewListener(r.setup.Identity.Address())
}

func makeMultiClientRole(setup RoleSetup, hub ConnHub, t *testing.T, stages int) multiClientRole {
	return multiClientRole{role: makeRole(setup, t, stages), hub: hub}
}

// NewPetra creates a new Proposer that executes the Petra protocol.
func NewPetra(setup RoleSetup, hub ConnHub, t *testing.T) *Petra {
	return &Petra{makeMultiClientRole(setup, hub, t, 5)}
}

// NewRobert creates a new Responder that executes the Robert protocol.
func NewRobert(setup RoleSetup, hub ConnHub, t *testing.T) *Robert {
	return &Robert{makeMultiClientRole(setup, hub, t, 5)}
}

// Execute executes the Petra protocol.
func (r *Petra) Execute(cfg ExecConfig) {
	assrt := assert.New(r.t)
	rng := rand.New(rand.NewSource(0x2994))

	prop := r.ChannelProposal(rng, &cfg)
	ch, err := r.ProposeChannel(prop)
	assrt.NoError(err)
	assrt.NotNil(ch)
	if err != nil {
		return
	}
	r.log.Infof("New Channel opened: %v", ch.Channel)

	// ignore proposal handler since Proposer doesn't accept any incoming channels
	_, wait := r.GoHandle(rng)

	// 1. channel controller set up
	r.waitStage()

	// 2. exchange final state
	ch.sendFinal()
	r.waitStage()

	// 3. Petra closes client, triggering disconnect and also closure of channel for Robert.
	assrt.NoError(r.Close())
	assrt.True(ch.IsClosed(), "closing client should close channel")
	wait() // Handle return
	r.waitStage()

	// 4. Robert closes client
	r.waitStage()

	r.assertPersistedPeerAndChannel(&cfg, ch.State())

	// 5. Restart clients and let them sync channels
	r.ReplaceClient()
	newCh := make(chan *paymentChannel, 1)
	r.OnNewChannel(func(_ch *paymentChannel) { newCh <- _ch })
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	r.Reconnect(ctx) // should connect to Robert
	select {
	case ch = <-newCh: // expected
		assrt.NotNil(ch)
	case <-time.After(r.timeout):
		r.t.Error("Expected channel to be restored")
	}

	r.waitStage()

	assrt.NoError(r.Close())
}

// Execute executes the Robert protocol.
func (r *Robert) Execute(cfg ExecConfig) {
	assrt := assert.New(r.t)
	rng := rand.New(rand.NewSource(0xB0B))

	waitListen := r.GoListen(r.setup.Listener)
	propHandler, waitHandler := r.GoHandle(rng)

	// receive one accepted proposal
	ch, err := propHandler.Next()
	assrt.NoError(err)
	assrt.NotNil(ch)
	if err != nil {
		return
	}
	r.log.Infof("New Channel opened: %v", ch.Channel)

	// 1st stage - channel controller set up
	r.waitStage()

	// 2. exchange final state
	ch.recvFinal()
	r.waitStage()

	// 3. Petra closed.
	r.waitStage()

	test.Within1s.Eventually(r.t, func(t test.T) {
		assert.True(t, ch.IsClosed(),
			"disconnecting Petra should eventually have closed channel")
		_, err = r.Channel(ch.ID())
		assert.Error(t, err, "disconnecting Petra should have removed channel")
	})

	// 4. Robert shuts down
	assrt.NoError(r.Close())
	waitListen()
	waitHandler()
	r.waitStage()

	r.assertPersistedPeerAndChannel(&cfg, ch.State())

	// 5. Restart clients and let them sync channels
	r.ReplaceClient()
	newCh := make(chan *paymentChannel, 1)
	r.OnNewChannel(func(_ch *paymentChannel) { newCh <- _ch })
	// Petra connects to us
	waitListen = r.GoListen(r.hub.NewListener(r.setup.Identity.Address()))
	defer waitListen()
	select {
	case ch = <-newCh: // expected
		assrt.NotNil(ch)
	case <-time.After(r.timeout):
		r.t.Error("Expected channel to be restored")
	}

	r.waitStage()

	assrt.NoError(r.Close())
}

func (r *multiClientRole) assertPersistedPeerAndChannel(cfg *ExecConfig, state *channel.State) {
	assrt := assert.New(r.t)
	_, them := r.Idxs(cfg.PeerAddrs)
	ps, err := r.setup.PR.ActivePeers(nil) // it should be a test persister, so no context needed
	peerAddr := cfg.PeerAddrs[them]
	assrt.NoError(err)
	assrt.Contains(ps, peerAddr)
	if len(ps) == 0 {
		return
	}
	chIt, err := r.setup.PR.RestorePeer(peerAddr)
	assrt.NoError(err)
	assrt.True(chIt.Next(nil))
	restoredCh := chIt.Channel()
	assrt.NoError(chIt.Close())
	assrt.Equal(restoredCh.ID(), state.ID)
	assrt.NoError(restoredCh.CurrentTXV.State.Equal(state))
}
