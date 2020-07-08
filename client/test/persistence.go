// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"context"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
)

type (
	multiClientRole struct {
		role
	}

	// Petra is the Proposer in a Persistence test.
	Petra struct{ multiClientRole }

	// Robert is the Responder in a persistence test.
	Robert struct{ multiClientRole }
)

// ReplaceClient replaces the client instance of the Role. Useful for
// persistence testing.
func (r *multiClientRole) ReplaceClient() {
	cl, err := client.New(r.setup.Identity, r.setup.Bus, r.setup.Funder, r.setup.Adjudicator, r.setup.Wallet)
	if err != nil {
		r.t.Fatal("Error recreating Client: ", err)
	}
	r.setClient(cl)
}

func makeMultiClientRole(setup RoleSetup, t *testing.T, stages int) multiClientRole {
	return multiClientRole{role: makeRole(setup, t, stages)}
}

// NewPetra creates a new Proposer that executes the Petra protocol.
func NewPetra(setup RoleSetup, t *testing.T) *Petra {
	return &Petra{makeMultiClientRole(setup, t, 6)}
}

// NewRobert creates a new Responder that executes the Robert protocol.
func NewRobert(setup RoleSetup, t *testing.T) *Robert {
	return &Robert{makeMultiClientRole(setup, t, 6)}
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

	// 2. send transfer
	ch.sendTransfer(big.NewInt(1), "send#1")
	r.waitStage()

	// 3. Both close client
	assrt.NoError(r.Close())
	assrt.True(ch.IsClosed(), "closing client should close channel")
	wait() // Handle return
	r.waitStage()

	r.assertPersistedPeerAndChannel(&cfg, ch.State())

	// 4. Restart clients and let them sync channels
	r.ReplaceClient()
	newCh := make(chan *paymentChannel, 1)
	r.OnNewChannel(func(_ch *paymentChannel) { newCh <- _ch })
	// ignore proposal handler
	_, wait = r.GoHandle(rng)
	defer wait()

	// 5. Clients restarted
	r.waitStage()

	// Restore channels locally
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	assrt.NoError(r.Restore(ctx)) // should restore channels
	select {
	case ch = <-newCh: // expected
		assrt.NotNil(ch)
	case <-time.After(r.timeout):
		r.t.Error("Expected channel to be restored")
	}

	r.waitStage()

	// 6. Finalize restored channel
	ch.recvFinal()

	assrt.NoError(r.Close())
}

// Execute executes the Robert protocol.
func (r *Robert) Execute(cfg ExecConfig) {
	assrt := assert.New(r.t)
	rng := rand.New(rand.NewSource(0xB0B))

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

	// 2. recv transfer
	ch.recvTransfer(big.NewInt(1), "recv#1")
	r.waitStage()

	// 3. Both close client
	assrt.NoError(r.Close())
	assrt.True(ch.IsClosed(), "closing client should close channel")
	waitHandler()
	r.waitStage()

	r.assertPersistedPeerAndChannel(&cfg, ch.State())

	// 4. Restart clients and let them sync channels
	r.ReplaceClient()
	newCh := make(chan *paymentChannel, 1)
	r.OnNewChannel(func(_ch *paymentChannel) { newCh <- _ch })

	// 5. Clients restarted
	r.waitStage()

	// Restore channels locally
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	assrt.NoError(r.Restore(ctx)) // should restore channels
	select {
	case ch = <-newCh: // expected
		assrt.NotNil(ch)
	case <-time.After(r.timeout):
		r.t.Error("Expected channel to be restored")
	}

	r.waitStage()

	// 6. Finalize restored channel
	ch.sendFinal()

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
