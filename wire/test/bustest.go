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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/test"
)

// timeout testNoReceive sub-test.
const testNoReceiveTimeout = 10 * time.Millisecond

// GenericBusTest tests the general functionality of a bus in the happy case: it
// tests that messages sent over the bus arrive at the correct destination. The
// parameter numClients controls how many clients communicate over the bus, and
// numMsgs controls how many messages each client sends to all other clients.
// The parameter busAssigner is used to assign a bus to each client, and must
// perform any necessary work to make clients able to communicate with each
// other (such as setting up dialers and listeners, in case of networking).
func GenericBusTest(t *testing.T, busAssigner func(wire.Account) wire.Bus, numClients, numMsgs int) {
	t.Helper()
	require.Greater(t, numClients, 1)
	require.Greater(t, numMsgs, 0)

	rng := test.Prng(t)
	type Client struct {
		r   *wire.Relay
		bus wire.Bus
		id  wire.Account
	}

	clients := make([]Client, numClients)
	for i := range clients {
		clients[i].r = wire.NewRelay()
		clients[i].id = wallettest.NewRandomAccount(rng)
		clients[i].bus = busAssigner(clients[i].id)
	}

	// Here, we have common, reused code.

	testNoReceive := func(t *testing.T) {
		t.Helper()
		ct := test.NewConcurrent(t)
		ctx, cancel := context.WithTimeout(context.Background(), testNoReceiveTimeout)
		defer cancel()
		for i := range clients {
			i := i
			go ct.StageN("receive timeout", numClients, func(t test.ConcT) {
				r := wire.NewReceiver()
				defer r.Close()
				err := clients[i].r.Subscribe(r, func(e *wire.Envelope) bool { return true })
				require.NoError(t, err)
				_, err = r.Next(ctx)
				require.Error(t, err)
			})
		}
		ct.Wait("receive timeout")
	}

	testPublishAndReceive := func(t *testing.T, waiting func()) {
		t.Helper()
		ct := test.NewConcurrent(t)
		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration((numClients)*(numClients-1)*numMsgs)*10*time.Millisecond)
		defer cancel()
		waiting()
		for sender := range clients {
			for recipient := range clients {
				if sender == recipient {
					continue
				}
				sender, recipient := sender, recipient
				origEnv := &wire.Envelope{
					Sender:    clients[sender].id.Address(),
					Recipient: clients[recipient].id.Address(),
					Msg:       wire.NewPingMsg(),
				}
				// Only subscribe to the current sender.
				recv := wire.NewReceiver()
				err := clients[recipient].r.Subscribe(recv, func(e *wire.Envelope) bool {
					return e.Sender.Equals(clients[sender].id.Address())
				})
				require.NoError(t, err)

				go ct.StageN("receive", numClients*(numClients-1), func(t test.ConcT) {
					defer recv.Close()
					for i := 0; i < numMsgs; i++ {
						e, err := recv.Next(ctx)
						require.NoError(t, err)
						require.Equal(t, e, origEnv)
					}
				})
				go ct.StageN("publish", numClients*(numClients-1), func(t test.ConcT) {
					for i := 0; i < numMsgs; i++ {
						err := clients[sender].bus.Publish(ctx, origEnv)
						require.NoError(t, err)
					}
				})
			}
		}
		ct.Wait("publish", "receive")

		// There must be no additional messages received.
		testNoReceive(t)
	}

	// Here, the actual test starts.
	// All following sub-tests operate on the same clients and subscriptions, so
	// changes made by one test are visible in the next tests.

	// First, we test that receiving without subscription will not result in any
	// messages.
	testNoReceive(t)
	// Then, we test that messages are received even if we subscribe after
	// publishing.
	testPublishAndReceive(t, func() {
		for i := range clients {
			err := clients[i].bus.SubscribeClient(clients[i].r, clients[i].id.Address())
			require.NoError(t, err)
		}
	})

	// Now that the subscriptions are already set up, we test that published
	// messages will be received if the subscription was in place before
	// publishing.
	testPublishAndReceive(t, func() {})
}
