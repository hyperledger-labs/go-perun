// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Proposer is a test client role. He proposes the new channel.
type Proposer struct {
	role
}

// NewProposer creates a new party that executes the Proposer protocol.
func NewProposer(setup RoleSetup, t *testing.T, numStages int) *Proposer {
	return &Proposer{role: makeRole(setup, t, numStages)}
}

// Execute executes the Proposer protocol.
func (r *Proposer) Execute(cfg ExecConfig, exec func(ExecConfig, *paymentChannel)) {
	rng := rand.New(rand.NewSource(0x471CE))
	assert := assert.New(r.t)

	prop := r.ChannelProposal(rng, &cfg)
	ch, err := r.ProposeChannel(prop)
	assert.NoError(err)
	assert.NotNil(ch)
	if err != nil {
		return
	}
	r.log.Infof("New Channel opened: %v", ch.Channel)

	// ignore proposal handler since Proposer doesn't accept any incoming channels
	_, wait := r.GoHandle(rng)
	defer wait()

	exec(cfg, ch)

	ch.Close() // May or may not already be closed due to channelConn closing.
	assert.NoError(r.Close())
}
