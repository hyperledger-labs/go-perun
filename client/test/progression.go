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
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
)

// ProgressionExecConfig contains config parameters for the progression test.
type ProgressionExecConfig struct {
	BaseExecConfig
}

// Watcher is a client that handles adjudicator events.
type Watcher struct {
	log.Logger
	registered chan *channel.RegisteredEvent
	progressed chan *channel.ProgressedEvent
}

func makeWatcher(log log.Logger) Watcher {
	return Watcher{
		Logger:     log,
		registered: make(chan *channel.RegisteredEvent),
		progressed: make(chan *channel.ProgressedEvent),
	}
}

// HandleAdjudicatorEvent is the callback for adjudicator event handling.
func (w *Watcher) HandleAdjudicatorEvent(e channel.AdjudicatorEvent) {
	w.Infof("HandleAdjudicatorEvent: %v", e)
	switch e := e.(type) {
	case *channel.RegisteredEvent:
		w.registered <- e
	case *channel.ProgressedEvent:
		w.progressed <- e
	}
}

// ----------------- BEGIN PAUL -----------------

// Paul is a test client role. He proposes the new channel.
type Paul struct {
	Proposer
	Watcher
}

// NewPaul creates a new party that executes the Paul protocol.
func NewPaul(t *testing.T, setup RoleSetup) *Paul {
	p := NewProposer(setup, t, 1)
	return &Paul{
		Proposer: *p,
		Watcher:  makeWatcher(p.log),
	}
}

// Execute executes the Paul protocol.
func (r *Paul) Execute(cfg ExecConfig) {
	r.Proposer.Execute(cfg, r.exec)
}

func (r *Paul) exec(_cfg ExecConfig, ch *paymentChannel) {
	assert := assert.New(r.t)
	ctx := r.Ctx()
	assetIdx := 0

	// start watcher
	go func() {
		r.log.Info("Starting channel watcher.")
		err := ch.Watch(r)
		r.log.Infof("Channel watcher returned: %v", err)
	}()

	r.waitStage() // wait for setup complete

	// progress
	assert.NoError(ch.ProgressBy(ctx, func(s *channel.State) {
		bal := func(user channel.Index) int64 {
			return s.Balances[assetIdx][user].Int64()
		}
		half := (bal(0) + bal(1)) / 2
		s.Balances[assetIdx][0] = big.NewInt(half)
		s.Balances[assetIdx][1] = big.NewInt(half)
	}))

	// await our progression confirmation
	<-r.progressed
	r.t.Logf("%v received progression confirmation 1", r.setup.Name)

	// await them progressing
	progEvent := <-r.progressed
	r.t.Logf("%v received progression confirmation 2", r.setup.Name)

	// await ready to conclude
	assert.NoError(progEvent.Timeout().Wait(ctx), "waiting for progression timeout")

	// withdraw
	assert.NoError(ch.Settle(ctx, false))
}

// ----------------- BEGIN PAULA -----------------

// Paula is a test client role. She proposes the new channel.
type Paula struct {
	Responder
	Watcher
}

// NewPaula creates a new party that executes the Paula protocol.
func NewPaula(t *testing.T, setup RoleSetup) *Paula {
	r := NewResponder(setup, t, 1)
	return &Paula{
		Responder: *r,
		Watcher:   makeWatcher(r.log),
	}
}

// Execute executes the Paula protocol.
func (r *Paula) Execute(cfg ExecConfig) {
	r.Responder.Execute(cfg, r.exec)
}

func (r *Paula) exec(_cfg ExecConfig, ch *paymentChannel, _ *acceptNextPropHandler) {
	assert := assert.New(r.t)
	ctx := r.Ctx()
	assetIdx := 0

	// start watcher
	go func() {
		r.log.Info("Starting channel watcher.")
		err := ch.Watch(r)
		r.log.Infof("Channel watcher returned: %v", err)
	}()

	r.waitStage() // wait for setup complete

	// await them registering
	regEvent := <-r.registered // get next, the same below
	_ = regEvent

	// await them progressing
	progEvent := <-r.progressed
	_ = progEvent
	r.t.Logf("%v received progression confirmation 1", r.setup.Name)

	// we progress
	assert.NoError(ch.ProgressBy(ctx, func(s *channel.State) {
		bal := func(user channel.Index) int64 {
			return s.Balances[assetIdx][user].Int64()
		}
		half := (bal(0) + bal(1)) / 2
		s.Balances[assetIdx][0] = big.NewInt(half + 10)
		s.Balances[assetIdx][1] = big.NewInt(half - 10)
	}))

	// await our progression confirmation
	progEvent = <-r.progressed // await progressed
	r.t.Logf("%v received progression confirmation 2", r.setup.Name)

	// await ready to conclude
	assert.NoError(progEvent.Timeout().Wait(ctx), "waiting for progression timeout")

	// withdraw
	assert.NoError(ch.Settle(ctx, true))
}
