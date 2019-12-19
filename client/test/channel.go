// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

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

// A paymentChannel is a wrapper around a client.Channel that implements payment
// logic and testing capabilities. It also implements the UpdateHandler
// interface so it can be its own update handler.
type paymentChannel struct {
	*client.Channel
	r *Role // Reuse of timeout and testing obj

	log     log.Logger
	handler chan bool
	err     chan error

	bals []channel.Bal // independent tracking of channel balance for testing
}

func newPaymentChannel(ch *client.Channel, r *Role) *paymentChannel {
	bals := make([]channel.Bal, 2)
	for i, balv := range ch.State().OfParts {
		bals[i] = new(big.Int).Set(balv[0])
	}

	return &paymentChannel{
		Channel: ch,
		r:       r,
		log:     r.log.WithField("channel", ch.ID()),
		handler: make(chan bool, 1),
		err:     make(chan error),
		bals:    bals,
	}
}

func (ch *paymentChannel) sendUpdate(update func(*channel.State), desc string) {
	ch.log.Debugf("Sending update: %s", desc)
	ctx, cancel := context.WithTimeout(context.Background(), ch.r.timeout)
	defer cancel()

	state := ch.State().Clone()
	update(state)
	state.Version++

	err := ch.Update(ctx, client.ChannelUpdate{
		State:    state,
		ActorIdx: ch.Idx(),
	})
	ch.log.Infof("Sent update: %s, err: %v", desc, err)
	assert.NoError(ch.r.t, err)
}

func (ch *paymentChannel) sendTransfer(amount channel.Bal, desc string) {
	ch.sendUpdate(
		func(state *channel.State) {
			transferBal(stateBals(state), ch.Idx(), amount)
		}, desc)

	transferBal(ch.bals, ch.Idx(), amount)
	ch.assertBals()
}

func (ch *paymentChannel) recvUpdate(accept bool, desc string) {
	ch.log.Debugf("Receiving update: %s, accept: %t", desc, accept)
	ch.handler <- accept

	var err error
	select {
	case err = <-ch.err:
		ch.log.Infof("Received update: %s, err: %v", desc, err)
	case <-time.After(ch.r.timeout):
		ch.r.t.Error("timeout: expected incoming channel update")
	}
	assert.NoError(ch.r.t, err)
}

func (ch *paymentChannel) recvTransfer(amount channel.Bal, desc string) {
	ch.recvUpdate(true, desc)
	transferBal(ch.bals, ch.Idx()^1, amount)
	ch.assertBals()
}

func (ch *paymentChannel) assertBals() {
	bals := stateBals(ch.State())
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
	ch.assertBals()
}

func (ch *paymentChannel) ListenUpdates() {
	ch.Channel.ListenUpdates(ch)
}

// The payment channel is its own update handler
func (ch *paymentChannel) Handle(up client.ChannelUpdate, res *client.UpdateResponder) {
	ch.log.Infof("Incoming channel update: %v", up)
	ctx, cancel := context.WithTimeout(context.Background(), ch.r.timeout)
	defer cancel()

	accept := <-ch.handler
	if accept {
		ch.log.Debug("Accepting...")
		ch.err <- res.Accept(ctx)
	} else {
		ch.log.Debug("Rejecting...")
		ch.err <- res.Reject(ctx, "Rejection")
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
	return []channel.Bal{state.OfParts[0][0], state.OfParts[1][0]}
}
