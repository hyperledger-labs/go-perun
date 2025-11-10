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

package simple

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	perunnet "perun.network/go-perun/wire/net"
	perunio "perun.network/go-perun/wire/perunio/serializer"
	wiretest "perun.network/go-perun/wire/test"
)

func TestBus(t *testing.T) {
	const numClients = 4
	const numMsgs = 5
	const defaultTimeout = 1000 * time.Millisecond

	commonName := "127.0.0.1"
	sans := []string{"127.0.0.1", "localhost"}
	tlsConfigs, err := generateSelfSignedCertConfigs(commonName, sans, numClients)
	assert.NoError(t, err)

	hosts := make([]string, numClients)
	for i := range numClients {
		port, err := findFreePort()
		assert.NoError(t, err)
		hosts[i] = fmt.Sprintf("127.0.0.1:%d", port)
	}

	dialers := make([]*Dialer, numClients)
	for j := range numClients {
		dialers[j] = NewTCPDialer(defaultTimeout, tlsConfigs[j])
		defer dialers[j].Close()
	}

	i := 0

	listeners := make([]*Listener, numClients)
	for j := range numClients {
		listener, err := NewTCPListener(hosts[j], tlsConfigs[j])
		assert.NoError(t, err)
		listeners[j] = listener
		defer listeners[j].Close()
	}

	wiretest.GenericBusTest(t, func(acc map[wallet.BackendID]wire.Account) (wire.Bus, wire.Bus) {
		for j := range numClients {
			dialers[j].Register(
				wire.AddressMapfromAccountMap(acc),
				hosts[i],
			)
		}

		bus := perunnet.NewBus(acc, dialers[i], perunio.Serializer())
		go bus.Listen(listeners[i])
		i++
		return bus, bus
	}, numClients, numMsgs)
}

func findFreePort() (int, error) {
	// Create a listener on a random port to get an available port.
	l, err := net.Listen("tcp", ":0") //nolint:gosec // Use ":0" to bind to a random free port
	if err != nil {
		return 0, err
	}
	defer l.Close()

	// Get the port from the listener's address
	addr := l.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
