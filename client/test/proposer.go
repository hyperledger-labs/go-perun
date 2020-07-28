// Copyright 2020 - See NOTICE file for copyright holders.
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
	"testing"

	"github.com/stretchr/testify/assert"
	pkgtest "perun.network/go-perun/pkg/test"
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
	rng := pkgtest.Prng(r.t, "proposer")
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
