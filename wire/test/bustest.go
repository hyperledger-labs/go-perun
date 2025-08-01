// Copyright 2025 - See NOTICE file for copyright holders.
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

	"perun.network/go-perun/wallet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/test"
)

// timeout testNoReceive sub-test.
const (
	testNoReceiveTimeout = 10 * time.Millisecond
	TestBackendID        = 0
)

// GenericBusTest tests the general functionality of a bus in the happy case: it
// tests that messages sent over the bus arrive at the correct destination. The
// parameter numClients controls how many clients communicate over the bus, and
// numMsgs controls how many messages each client sends to all other clients.
// The parameter busAssigner is used to assign a bus to each client, and must
// perform any necessary work to make clients able to communicate with each
// other (such as setting up dialers and listeners, in case of networking). It
// can either return the same bus twice, or separately select a bus to subscribe
// the client to, and a bus the client should use for publishing messages.
func GenericBusTest(t *testing.T,
	busAssigner func(map[wallet.BackendID]wire.Account) (pub wire.Bus, sub wire.Bus),
	numClients, numMsgs int,
) {
	t.Helper()
	require.Greater(t, numClients, 1)
	require.Positive(t, numMsgs)

	rng := test.Prng(t)

	type Client struct {
		r        *wire.Relay
		pub, sub wire.Bus
		id       map[wallet.BackendID]wire.Account
	}

	clients := make([]Client, numClients)
	for i := range clients {
		clients[i].r = wire.NewRelay()
		clients[i].id = NewRandomAccountMap(rng, TestBackendID)
		clients[i].pub, clients[i].sub = busAssigner(clients[i].id)
	}

	// Here, we have common, reused code.

	testNoReceive := func(t *testing.T) {
		t.Helper()
		ct := test.NewConcurrent(t)

		ctx, cancel := context.WithTimeout(context.Background(), testNoReceiveTimeout)
		defer cancel()

		for i := range clients {
			go ct.StageN("receive timeout", numClients, func(t test.ConcT) {
				r := wire.NewReceiver()
				defer r.Close()
				err := clients[i].r.Subscribe(r, func(e *wire.Envelope) bool { return true })
				assert.NoError(t, err)
				_, err = r.Next(ctx)
				assert.Error(t, err)
			})
		}

		ct.Wait("receive timeout")
	}

	testPublishAndReceive := func(t *testing.T, waiting func()) {
		t.Helper()
		ct := test.NewConcurrent(t)

		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration((numClients)*(numClients-1)*numMsgs)*100*time.Millisecond)
		defer cancel()

		waiting()

		for sender := range clients {
			for recipient := range clients {
				if sender == recipient {
					continue
				}
				origEnv := &wire.Envelope{
					Sender:    wire.AddressMapfromAccountMap(clients[sender].id),
					Recipient: wire.AddressMapfromAccountMap(clients[recipient].id),
					Msg:       wire.NewPingMsg(),
				}
				// Only subscribe to the current sender.
				recv := wire.NewReceiver()
				err := clients[recipient].r.Subscribe(recv, func(e *wire.Envelope) bool {
					return equalMaps(e.Sender, wire.AddressMapfromAccountMap(clients[sender].id))
				})
				require.NoError(t, err)

				go ct.StageN("receive", numClients*(numClients-1), func(t test.ConcT) {
					defer recv.Close()

					for range numMsgs {
						e, err := recv.Next(ctx)
						assert.NoError(t, err)
						assert.Equal(t, e, origEnv)
					}
				})
				go ct.StageN("publish", numClients*(numClients-1), func(t test.ConcT) {
					for range numMsgs {
						err := clients[sender].pub.Publish(ctx, origEnv)
						assert.NoError(t, err)
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
			err := clients[i].sub.SubscribeClient(clients[i].r, wire.AddressMapfromAccountMap(clients[i].id))
			require.NoError(t, err)
		}
	})

	// Now that the subscriptions are already set up, we test that published
	// messages will be received if the subscription was in place before
	// publishing.
	testPublishAndReceive(t, func() {})
}

func equalMaps(a, b map[wallet.BackendID]wire.Address) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if !v.Equal(b[k]) {
			return false
		}
	}
	return true
}
