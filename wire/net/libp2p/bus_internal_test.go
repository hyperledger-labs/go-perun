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

package libp2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	perunnet "perun.network/go-perun/wire/net"
	perunio "perun.network/go-perun/wire/perunio/serializer"
	wiretest "perun.network/go-perun/wire/test"
)

func TestBus(t *testing.T) {
	const numClients = 4
	const numMsgs = 4

	i := 0 // Keep track of the current client ID
	dialers := make([]*Dialer, numClients)
	accs := make([]*Account, numClients)
	defer func() {
		// Close all accounts
		for _, acc := range accs {
			if acc != nil {
				acc.Close()
			}
		}
	}()

	wiretest.GenericBusTest(t, func(acc map[wallet.BackendID]wire.Account) (wire.Bus, wire.Bus) {
		libp2pAcc, ok := acc[channel.TestBackendID].(*Account)
		assert.True(t, ok)

		accs[i] = libp2pAcc
		dialer := NewP2PDialer(libp2pAcc)
		dialers[i] = dialer

		if i == numClients-1 {
			// Register all peers
			for j := range numClients {
				for k := range numClients {
					if j != k {
						peerAcc := accs[k]
						dialers[j].Register(
							map[wallet.BackendID]wire.Address{
								channel.TestBackendID: peerAcc.Address(),
							},
							peerAcc.ID().String())
					}
				}
			}
		}

		listener := NewP2PListener(libp2pAcc)

		bus := perunnet.NewBus(acc, dialer, perunio.Serializer())
		go bus.Listen(listener)

		i++

		return bus, bus
	}, numClients, numMsgs)
}
