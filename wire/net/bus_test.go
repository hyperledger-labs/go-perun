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

package net_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"perun.network/go-perun/wallet"

	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/net"
	nettest "perun.network/go-perun/wire/net/test"
	perunio "perun.network/go-perun/wire/perunio/serializer"
	wiretest "perun.network/go-perun/wire/test"
)

func TestBus(t *testing.T) {
	const numClients = 16
	const numMsgs = 16

	var hub nettest.ConnHub

	wiretest.GenericBusTest(t, func(acc map[wallet.BackendID]wire.Account) (wire.Bus, wire.Bus) {
		bus := net.NewBus(acc, hub.NewNetDialer(), perunio.Serializer())
		hub.OnClose(func() { bus.Close() })
		go bus.Listen(hub.NewNetListener(wire.AddressMapfromAccountMap(acc)))
		return bus, bus
	}, numClients, numMsgs)

	require.NoError(t, hub.Close())
}
