// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/test"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
)

// GenericBusTest tests the general functionality of a bus in the happy case: it
// tests that messages sent over the bus arrive at the correct destination. The
// parameter numClients controls how many clients communicate over the bus, and
// numMsgs controls how many messages each client sends to all other clients.
func GenericBusTest(t *testing.T, bus wire.Bus, numClients, numMsgs int) {
	rng := test.Prng(t)
	type Client struct {
		r  *wire.Relay
		id wire.Account
	}

	clients := make([]Client, numClients)
	for i := range clients {
		clients[i].r = wire.NewRelay()
		clients[i].id = wallettest.NewRandomAccount(rng)
	}

	// Here, we have common, reused code.

	testNoReceive := func(t *testing.T) {
		ct := test.NewConcurrent(t)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		for i := range clients {
			i := i
			go ct.StageN("receive timeout", numClients, func(t require.TestingT) {
				r := wire.NewReceiver()
				defer r.Close()
				clients[i].r.Subscribe(r, func(e *wire.Envelope) bool { return true })
				_, err := r.Next(ctx)
				require.Error(t, err)
			})
		}
		ct.Wait("receive timeout")
	}

	testPublishAndReceive := func(t *testing.T, waiting func()) {
		ct := test.NewConcurrent(t)
		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(numClients*numClients*numMsgs)*10*time.Millisecond)
		defer cancel()
		for sender := range clients {
			for recipient := range clients {
				sender, recipient := sender, recipient
				origEnv := &wire.Envelope{
					Sender:    clients[sender].id.Address(),
					Recipient: clients[recipient].id.Address(),
					Msg:       wire.NewPingMsg(),
				}
				// Only subscribe to the current sender.
				recv := wire.NewReceiver()
				clients[recipient].r.Subscribe(recv, func(e *wire.Envelope) bool {
					return e.Sender.Equals(clients[sender].id.Address())
				})

				go ct.StageN("receive", numClients*numClients, func(t require.TestingT) {
					defer recv.Close()
					for i := 0; i < numMsgs; i++ {
						e, err := recv.Next(ctx)
						require.NoError(t, err)
						require.Same(t, e, origEnv)
					}
				})
				go ct.StageN("publish", numClients*numClients, func(t require.TestingT) {
					for i := 0; i < numMsgs; i++ {
						err := bus.Publish(ctx, origEnv)
						require.NoError(t, err)
					}
				})
			}
		}
		waiting()
		ct.Wait("publish", "receive")

		// There must be no additional messages received.
		testNoReceive(t)
	}

	// Here, the actual test starts.
	// All following sub-tests operate on the same clients and subscriptions, so
	// changes made by one test are visible in the next tests.

	// First, we test that receiving without subscription will not result in any
	// messages.
	t.Run("empty receive", testNoReceive)
	// Then, we test that messages are received even if we subscribe after
	// publishing.
	t.Run("subscribe after publish", func(t *testing.T) {
		testPublishAndReceive(t, func() {
			for i := range clients {
				err := bus.SubscribeClient(clients[i].r, clients[i].id.Address())
				require.NoError(t, err)
			}
		})
	})
	// Now that the subscriptions are already set up, we test that published
	// messages will be received if the subscription was in place before
	// publishing.
	t.Run("normal publish", func(t *testing.T) {
		testPublishAndReceive(t, func() {})
	})
}
