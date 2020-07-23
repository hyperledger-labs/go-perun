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
	"math/rand"
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

	propHandler, waitHandler := r.GoHandle(rng)
	defer waitHandler()

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
