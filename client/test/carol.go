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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Carol is a Responder. She accepts an incoming channel proposal.
type Carol struct {
	Responder
}

// NewCarol creates a new Responder that executes the Carol protocol.
func NewCarol(setup RoleSetup, t *testing.T) *Carol {
	return &Carol{Responder: *NewResponder(setup, t, 3)}
}

// Execute executes the Carol protocol.
func (r *Carol) Execute(cfg ExecConfig) {
	r.Responder.Execute(cfg, r.exec)
}

func (r *Carol) exec(_cfg ExecConfig, ch *paymentChannel) {
	cfg := _cfg.(*MalloryCarolExecConfig)
	assert := assert.New(r.t)
	_, them := r.Idxs(cfg.PeerAddrs())

	// start watcher
	watcher := make(chan error)
	go func() {
		r.log.Info("Starting channel watcher.")
		watcher <- ch.Watch()
		r.log.Debug("Channel watcher returned.")
	}()

	// 1st stage - channel controller set up
	r.waitStage()

	// Carol receives some updates from Mallory
	for i := 0; i < cfg.NumPayments[them]; i++ {
		ch.recvTransfer(cfg.TxAmounts[them], fmt.Sprintf("Mallory#%d", i))
	}
	// 2nd stage - txs received
	r.waitStage()

	r.log.Debug("Waiting for watcher to return...")
	select {
	case err := <-watcher:
		assert.NoError(err)
	case <-time.After(r.timeout):
		r.t.Error("expected watcher to return")
		return
	}

	// 3rd stage - channel settled
	r.waitStage()
}
