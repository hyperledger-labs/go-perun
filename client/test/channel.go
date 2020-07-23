// Copyright 2019 - See NOTICE file for copyright holders.
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
	"time"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/log"
)

type (
	// A paymentChannel is a wrapper around a client.Channel that implements payment
	// logic and testing capabilities. It also implements the UpdateHandler
	// interface so it can be its own update handler.
	paymentChannel struct {
		*client.Channel
		r *role // Reuse of timeout and testing obj

		log     log.Logger
		handler chan bool
		res     chan handlerRes

		bals []channel.Bal // independent tracking of channel balance for testing
	}

	// A handlerRes encapsulates the result of a channel handling request
	handlerRes struct {
		up  client.ChannelUpdate
		err error
	}
)

func newPaymentChannel(ch *client.Channel, r *role) *paymentChannel {
	return &paymentChannel{
		Channel: ch,
		r:       r,
		log:     r.log.WithField("channel", ch.ID()),
		handler: make(chan bool, 1),
		res:     make(chan handlerRes),
		bals:    channel.CloneBals(stateBals(ch.State())),
	}
}

func (ch *paymentChannel) sendUpdate(update func(*channel.State), desc string) {
	ch.log.Debugf("Sending update: %s", desc)
	ctx, cancel := context.WithTimeout(context.Background(), ch.r.timeout)
	defer cancel()

	err := ch.UpdateBy(ctx, update)
	ch.log.Infof("Sent update: %s, err: %v", desc, err)
	assert.NoError(ch.r.t, err)
}

func (ch *paymentChannel) sendTransfer(amount channel.Bal, desc string) {
	ch.sendUpdate(
		func(state *channel.State) {
			transferBal(stateBals(state), ch.Idx(), amount)
		}, desc)

	transferBal(ch.bals, ch.Idx(), amount)
	ch.assertBals(ch.State())
}

func (ch *paymentChannel) recvUpdate(accept bool, desc string) *channel.State {
	ch.log.Debugf("Receiving update: %s, accept: %t", desc, accept)
	ch.handler <- accept

	select {
	case res := <-ch.res:
		ch.log.Infof("Received update: %s, err: %v", desc, res.err)
		assert.NoError(ch.r.t, res.err)
		return res.up.State
	case <-time.After(ch.r.timeout):
		ch.r.t.Error("timeout: expected incoming channel update")
		return nil
	}
}

func (ch *paymentChannel) recvTransfer(amount channel.Bal, desc string) {
	state := ch.recvUpdate(true, desc)
	if state != nil {
		transferBal(ch.bals, ch.Idx()^1, amount)
		ch.assertBals(state)
	} // else recvUpdate timed out
}

func (ch *paymentChannel) assertBals(state *channel.State) {
	bals := stateBals(state)
	ch.log.Infof(
		"Tracked balance: [ %v %v ], channel: [ %v %v ]",
		ch.bals[0], ch.bals[1],
		bals[0], bals[1],
	)
	assert := assert.New(ch.r.t)
	assert.Zerof(bals[0].Cmp(ch.bals[0]), "bal[0]: %v != %v", bals[0], ch.bals[0])
	assert.Zerof(bals[1].Cmp(ch.bals[1]), "bal[1]: %v != %v", bals[1], ch.bals[1])
}

func (ch *paymentChannel) sendFinal() {
	ch.sendUpdate(func(state *channel.State) {
		state.IsFinal = true
	}, "final")
	assert.True(ch.r.t, ch.State().IsFinal)
}

func (ch *paymentChannel) recvFinal() {
	ch.recvUpdate(true, "final")
	assert.True(ch.r.t, ch.State().IsFinal)
}

func (ch *paymentChannel) settleChan() {
	ctx, cancel := context.WithTimeout(context.Background(), ch.r.timeout)
	defer cancel()
	assert.NoError(ch.r.t, ch.Settle(ctx))
	ch.assertBals(ch.State())
}

// The payment channel is its own update handler
func (ch *paymentChannel) Handle(up client.ChannelUpdate, res *client.UpdateResponder) {
	ch.log.Infof("Incoming channel update: %v", up)
	ctx, cancel := context.WithTimeout(context.Background(), ch.r.timeout)
	defer cancel()

	accept := <-ch.handler
	if accept {
		ch.log.Debug("Accepting...")
		ch.res <- handlerRes{up, res.Accept(ctx)}
	} else {
		ch.log.Debug("Rejecting...")
		ch.res <- handlerRes{up, res.Reject(ctx, "Rejection")}
	}
}

func transferBal(bals []channel.Bal, ourIdx channel.Index, amount *big.Int) {
	a := new(big.Int).Set(amount) // local copy because we mutate it
	otherIdx := ourIdx ^ 1
	ourBal := bals[ourIdx]
	otherBal := bals[otherIdx]
	otherBal.Add(otherBal, a)
	ourBal.Add(ourBal, a.Neg(a))
}

func stateBals(state *channel.State) []channel.Bal {
	return state.Balances[0]
}
