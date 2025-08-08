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
)

// Responder is a test client role. He accepts an incoming channel proposal.
type Responder struct {
	role
}

// NewResponder creates a new party that executes the Responder protocol.
func NewResponder(t *testing.T, setup RoleSetup, numStages int) *Responder {
	t.Helper()
	return &Responder{role: makeRole(t, setup, numStages)}
}

// Execute executes the Responder protocol.
func (r *Responder) Execute(cfg ExecConfig, exec func(ExecConfig, *paymentChannel, *acceptNextPropHandler)) {
	rng := r.NewRng()

	propHandler, waitHandler := r.GoHandle(rng)

	defer func() {
		r.RequireNoError(r.Close())
		waitHandler()
	}()

	// receive one accepted proposal
	ch, err := propHandler.Next()
	r.RequireNoError(err)
	r.RequireTrue(ch != nil)
	if err != nil {
		return
	}

	r.log.Infof("New Channel opened: %v", ch.Channel)

	exec(cfg, ch, propHandler)

	r.RequireNoError(ch.Close())
}

// Errors returns the error channel.
func (r *Responder) Errors() <-chan error {
	return r.errs
}
