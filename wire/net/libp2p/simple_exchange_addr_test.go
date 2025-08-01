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

// This test uses the wire/net/libp2p implementation of Account and Address
// to test the default implementation of wire.
//
//nolint:testpackage
package libp2p

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"perun.network/go-perun/channel"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/wire"
	wirenet "perun.network/go-perun/wire/net"
	perunio "perun.network/go-perun/wire/perunio/serializer"
	wiretest "perun.network/go-perun/wire/test"
	ctxtest "polycry.pt/poly-go/context/test"
	"polycry.pt/poly-go/test"
)

const timeout = 100 * time.Millisecond

func TestExchangeAddrs_ConnFail(t *testing.T) {
	rng := test.Prng(t)
	a, _ := newPipeConnPair()
	a.Close()
	acc := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	defer acc[channel.TestBackendID].(*Account).Close()
	addr, err := wirenet.ExchangeAddrsPassive(context.Background(), acc, a)
	assert.Nil(t, addr)
	assert.Error(t, err)
}

func TestExchangeAddrs_Success(t *testing.T) {
	rng := test.Prng(t)
	conn0, conn1 := newPipeConnPair()
	defer conn0.Close()
	account0, account1 := wiretest.NewRandomAccountMap(rng, channel.TestBackendID), wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	defer account0[channel.TestBackendID].(*Account).Close()
	defer account1[channel.TestBackendID].(*Account).Close()
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer conn1.Close()

		recvAddr0, err := wirenet.ExchangeAddrsPassive(context.Background(), account1, conn1)
		assert.NoError(t, err)
		assert.True(t, channel.EqualWireMaps(recvAddr0, wire.AddressMapfromAccountMap(account0)))
	}()

	err := wirenet.ExchangeAddrsActive(context.Background(), account0, wire.AddressMapfromAccountMap(account1), conn0)
	require.NoError(t, err)

	wg.Wait()
}

func TestExchangeAddrs_BogusMsg(t *testing.T) {
	rng := test.Prng(t)
	acc := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	defer acc[channel.TestBackendID].(*Account).Close()
	conn := newMockConn()
	conn.recvQueue <- newRandomEnvelope(rng, wire.NewPingMsg())
	addr, err := wirenet.ExchangeAddrsPassive(context.Background(), acc, conn)

	assert.Error(t, err, "ExchangeAddrs should error when peer sends a non-AuthResponseMsg")
	assert.Nil(t, addr)
}

func TestExchangeAddrs_Timeout(t *testing.T) {
	rng := test.Prng(t)
	a, _ := newPipeConnPair()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ctxtest.AssertTerminates(t, 20*timeout, func() {
		acc := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
		defer acc[channel.TestBackendID].(*Account).Close()
		addr, err := wirenet.ExchangeAddrsPassive(ctx, acc, a)
		assert.Nil(t, addr)
		assert.Error(t, err)
	})
}

// newPipeConnPair creates endpoints that are connected via pipes.
func newPipeConnPair() (a wirenet.Conn, b wirenet.Conn) {
	c0, c1 := net.Pipe()
	ser := perunio.Serializer()
	return wirenet.NewIoConn(c0, ser), wirenet.NewIoConn(c1, ser)
}

// NewRandomEnvelope returns an envelope around message m with random sender and
// recipient generated using randomness from rng.
func newRandomEnvelope(rng *rand.Rand, m wire.Msg) *wire.Envelope {
	return &wire.Envelope{
		Sender:    NewRandomAddresses(rng),
		Recipient: NewRandomAddresses(rng),
		Msg:       m,
	}
}
