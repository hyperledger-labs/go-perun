// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Responder is a test client role. He accepts an incoming channel proposal.
type Responder struct {
	role
}

// NewResponder creates a new party that executes the Responder protocol.
func NewResponder(setup RoleSetup, t *testing.T, numStages int) *Responder {
	return &Responder{role: makeRole(setup, t, numStages)}
}

// Execute executes the Responder protocol.
func (r *Responder) Execute(cfg ExecConfig, exec func(ExecConfig, *paymentChannel)) {
	rng := rand.New(rand.NewSource(0xB0B))
	assert := assert.New(r.t)
	propHandler := r.AcceptAllPropHandler(rng)

	var listenWg sync.WaitGroup
	listenWg.Add(2)
	defer func() {
		r.log.Debug("Waiting for listeners to return...")
		listenWg.Wait()
	}()
	go func() {
		defer listenWg.Done()
		r.log.Info("Starting peer listener.")
		r.Listen(r.setup.Listener)
		r.log.Debug("Peer listener returned.")
	}()
	go func() {
		defer listenWg.Done()
		r.log.Info("Starting request handler.")
		r.Handle(propHandler, r.UpdateHandler())
		r.log.Debug("Request handler returned.")
	}()

	// receive one accepted proposal
	ch, err := propHandler.Next()
	assert.NoError(err)
	assert.NotNil(ch)
	if err != nil {
		return
	}

	r.log.Infof("New Channel opened: %v", ch.Channel)

	exec(cfg, ch)

	assert.NoError(ch.Close())
	assert.NoError(r.Close())
}
