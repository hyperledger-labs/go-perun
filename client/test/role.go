// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

// Package test contains helpers for testing the client
package test // import "perun.network/go-perun/client/test"

import (
	"context"
	"math/big"
	"math/rand"
	"sync"
	"testing"
	"time"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/log"
	"perun.network/go-perun/peer"
	wallettest "perun.network/go-perun/wallet/test"
)

type (
	// A Role is a client.Client together with a protocol execution path.
	Role struct {
		*client.Client
		setup RoleSetup
		// we use the Client as Closer
		timeout   time.Duration
		log       log.Logger
		t         *testing.T
		numStages int
		stages    Stages
	}

	// RoleSetup contains the injectables for setting up the client.
	RoleSetup struct {
		Name     string
		Identity peer.Identity
		Dialer   peer.Dialer
		Listener peer.Listener
		Funder   channel.Funder
		Settler  channel.Settler
		Timeout  time.Duration
	}

	// ExecConfig contains additional config parameters for the tests.
	ExecConfig struct {
		PeerAddrs       []peer.Address // must match RoleSetup.Identity of [Alice, Bob]
		Asset           channel.Asset  // single Asset to use in this channel
		InitBals        []*big.Int     // channel deposit of [Alice, Bob]
		NumUpdatesBob   int            // 1st Bob sends updates
		NumUpdatesAlice int            // then 2nd Alice sends updates
		TxAmountBob     *big.Int       // amount that Bob sends per udpate
		TxAmountAlice   *big.Int       // amount that Alice sends per udpate
	}

	// Stages are used to synchronize multiple roles.
	Stages = []sync.WaitGroup
)

// MakeRole creates a client for the given setup and wraps it into a Role.
func MakeRole(setup RoleSetup, propHandler client.ProposalHandler, t *testing.T, numStages int) Role {
	cl := client.New(setup.Identity, setup.Dialer, propHandler, setup.Funder, setup.Settler)
	return Role{
		Client:    cl,
		setup:     setup,
		timeout:   setup.Timeout,
		log:       cl.Log().WithField("role", setup.Name),
		t:         t,
		numStages: numStages,
	}
}

// EnableStages optionally enables the synchronization of this role at different
// stages of the Execute protocol. EnableStages should be called on a single
// role and the resulting slice set on all remaining roles by calling SetStages
// on them.
func (r *Role) EnableStages() Stages {
	r.stages = make(Stages, r.numStages)
	for i := range r.stages {
		r.stages[i].Add(1)
	}
	return r.stages
}

// SetStages optionally sets a slice of WaitGroup barriers to wait on at
// different stages of the Execute protocol. It should be created by any role by
// calling EnableStages().
func (r *Role) SetStages(st Stages) {
	if len(st) != r.numStages {
		panic("number of stages don't match")
	}

	r.stages = st
	for i := range r.stages {
		r.stages[i].Add(1)
	}
}

func (r *Role) waitStage() {
	if r.stages != nil {
		r.numStages--
		stage := &r.stages[r.numStages]
		stage.Done()
		stage.Wait()
	}
}

type (
	// acceptAllPropHandler is a channel proposal handler that accepts all channel
	// requests. It generates a random account for each channel.
	// Each accepted channel is put on the chans go channel.
	acceptAllPropHandler struct {
		chans   chan channelAndError
		log     log.Logger
		rng     *rand.Rand
		timeout time.Duration
	}

	// channelAndError bundles the return parameters of ProposalResponder.Accept
	// to be able to send them over a channel.
	channelAndError struct {
		channel *client.Channel
		err     error
	}
)

func newAcceptAllPropHandler(rng *rand.Rand, timeout time.Duration) *acceptAllPropHandler {
	return &acceptAllPropHandler{
		chans:   make(chan channelAndError),
		rng:     rng,
		timeout: timeout,
		log:     log.Get(), // default logger without fields
	}
}

func (h *acceptAllPropHandler) Handle(req *client.ChannelProposalReq, res *client.ProposalResponder) {
	h.log.Infof("Accepting incoming channel request: %v", req)
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	ch, err := res.Accept(ctx, client.ProposalAcc{
		Participant: wallettest.NewRandomAccount(h.rng),
	})
	h.chans <- channelAndError{ch, err}
}
