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
	"context"
	"math/big"
	"testing"
	"time"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wire"
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

const (
	defaultTestTimeout   = 10 * time.Second
	petraRobertNumStages = 6
)

// ReplaceClient replaces the client instance of the Role. Useful for
// persistence testing.
func (r *multiClientRole) ReplaceClient() {
	cl, err := client.New(r.setup.Identity.Address(), r.setup.Bus, r.setup.Funder, r.setup.Adjudicator, r.setup.Wallet, r.setup.Watcher)
	r.RequireNoErrorf(err, "recreating client")
	r.setClient(cl)
}

func makeMultiClientRole(t *testing.T, setup RoleSetup, stages int) multiClientRole {
	t.Helper()
	return multiClientRole{role: makeRole(t, setup, stages)}
}

// NewPetra creates a new Proposer that executes the Petra protocol.
func NewPetra(t *testing.T, setup RoleSetup) *Petra {
	t.Helper()
	return &Petra{makeMultiClientRole(t, setup, petraRobertNumStages)}
}

// NewRobert creates a new Responder that executes the Robert protocol.
func NewRobert(t *testing.T, setup RoleSetup) *Robert {
	t.Helper()
	return &Robert{makeMultiClientRole(t, setup, petraRobertNumStages)}
}

// Execute executes the Petra protocol.
func (r *Petra) Execute(cfg ExecConfig) {
	rng := r.NewRng()

	prop := r.LedgerChannelProposal(rng, cfg)
	ch, err := r.ProposeChannel(prop)
	r.RequireNoError(err)
	r.RequireTrue(ch != nil)
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
	r.RequireNoError(r.Close())
	r.RequireTruef(ch.IsClosed(), "closing client should close channel")
	wait() // Handle return
	r.waitStage()

	r.assertPersistedPeerAndChannel(cfg, ch.State())

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
	r.RequireNoError(r.Restore(ctx)) // should restore channels
	select {
	case ch = <-newCh: // expected
		r.RequireTrue(ch != nil)
	case <-time.After(r.timeout):
		r.RequireTruef(false, "expected channel to be restored")
	}

	r.waitStage()

	// 6. Finalize restored channel
	ch.recvFinal()

	ch.settle()

	r.RequireNoError(r.Close())
}

// Execute executes the Robert protocol.
func (r *Robert) Execute(cfg ExecConfig) {
	rng := r.NewRng()

	propHandler, waitHandler := r.GoHandle(rng)

	// receive one accepted proposal
	ch, err := propHandler.Next()
	r.RequireNoError(err)
	r.RequireTrue(ch != nil)
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
	r.RequireNoError(r.Close())
	r.RequireTruef(ch.IsClosed(), "closing client should close channel")
	waitHandler()
	r.waitStage()

	r.assertPersistedPeerAndChannel(cfg, ch.State())

	// 4. Restart clients and let them sync channels
	r.ReplaceClient()
	newCh := make(chan *paymentChannel, 1)
	r.OnNewChannel(func(_ch *paymentChannel) { newCh <- _ch })

	// 5. Clients restarted
	r.waitStage()

	// Restore channels locally
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	r.RequireNoError(r.Restore(ctx)) // should restore channels
	select {
	case ch = <-newCh: // expected
		r.RequireTrue(ch != nil)
	case <-time.After(r.timeout):
		r.RequireTruef(false, "Expected channel to be restored")
	}

	r.waitStage()

	// 6. Finalize restored channel
	ch.sendFinal()

	ch.settle()

	r.RequireNoError(r.Close())
}

func (r *multiClientRole) assertPersistedPeerAndChannel(cfg ExecConfig, state *channel.State) {
	_, them := r.Idxs(cfg.Peers())
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	ps, err := r.setup.PR.ActivePeers(ctx) // it should be a test persister, so no context needed
	peerAddr := cfg.Peers()[them]
	r.RequireNoError(err)
	r.RequireTrue(addresses(ps).contains(peerAddr))
	if len(ps) == 0 {
		return
	}
	chIt, err := r.setup.PR.RestorePeer(peerAddr)
	r.RequireNoError(err)
	r.RequireTrue(chIt.Next(ctx))
	restoredCh := chIt.Channel()
	r.RequireNoError(chIt.Close())
	r.RequireTrue(restoredCh.ID() == state.ID)
	r.RequireNoError(restoredCh.CurrentTXV.State.Equal(state))
}

// Errors returns the error channel.
func (r *multiClientRole) Errors() <-chan error {
	return r.errs
}

type addresses []wire.Address

func (a addresses) contains(b wire.Address) bool {
	for _, addr := range a {
		if addr.Equal(b) {
			return true
		}
	}
	return false
}
