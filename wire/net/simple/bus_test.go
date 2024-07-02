// Copyright 2024 - See NOTICE file for copyright holders.
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

package simple_test

import (
	"fmt"
	"testing"
	"time"

	"net"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/wire"
	perunnet "perun.network/go-perun/wire/net"
	"perun.network/go-perun/wire/net/simple"
	perunio "perun.network/go-perun/wire/perunio/serializer"
	wiretest "perun.network/go-perun/wire/test"
)

func TestBus(t *testing.T) {
	const numClients = 16
	const numMsgs = 16
	const defaultTimeout = 15 * time.Millisecond

	commonName := "127.0.0.1"
	sans := []string{"127.0.0.1", "localhost"}
	tlsConfigs, err := simple.GenerateSelfSignedCertConfigs(commonName, sans, numClients)
	assert.NoError(t, err)

	hosts := make([]string, numClients)
	for i := 0; i < numClients; i++ {
		port, err := findFreePort()
		assert.NoError(t, err)
		hosts[i] = fmt.Sprintf("127.0.0.1:%d", port)
	}

	dialers := make([]*simple.Dialer, numClients)
	for j := 0; j < numClients; j++ {
		dialers[j] = simple.NewTCPDialer(defaultTimeout, tlsConfigs[j])
	}

	i := 0

	wiretest.GenericBusTest(t, func(acc wire.Account) (wire.Bus, wire.Bus) {
		for j := 0; j < numClients; j++ {
			dialers[j].Register(acc.Address(), hosts[i])
		}

		bus := perunnet.NewBus(acc, dialers[i], perunio.Serializer())
		listener, err := simple.NewTCPListener(hosts[i], tlsConfigs[i])
		assert.NoError(t, err)
		go bus.Listen(listener)
		i++
		return bus, bus
	}, numClients, numMsgs)

	for j := 0; j < numClients; j++ {
		assert.NoError(t, dialers[j].Close())
	}
}

func findFreePort() (int, error) {
	// Create a listener on a random port to get an available port.
	l, err := net.Listen("tcp", ":0") // Use ":0" to bind to a random free port
	if err != nil {
		return 0, err
	}
	defer l.Close()

	// Get the port from the listener's address
	addr := l.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
