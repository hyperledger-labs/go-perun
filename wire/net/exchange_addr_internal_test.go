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

package net

import (
	"context"
	"sync"
	"testing"

	"perun.network/go-perun/channel"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/wire"
	wiretest "perun.network/go-perun/wire/test"
	ctxtest "polycry.pt/poly-go/context/test"
	"polycry.pt/poly-go/test"
)

func TestExchangeAddrs_ConnFail(t *testing.T) {
	rng := test.Prng(t)
	a, _ := newPipeConnPair()
	a.Close()
	addr, err := ExchangeAddrsPassive(context.Background(), wiretest.NewRandomAccountMap(rng, channel.TestBackendID), a)
	assert.Nil(t, addr)
	assert.Error(t, err)
}

func TestExchangeAddrs_Success(t *testing.T) {
	rng := test.Prng(t)
	conn0, conn1 := newPipeConnPair()
	defer conn0.Close()
	account0, account1 := wiretest.NewRandomAccountMap(rng, channel.TestBackendID), wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer conn1.Close()

		recvAddr0, err := ExchangeAddrsPassive(context.Background(), account1, conn1)
		assert.NoError(t, err)
		assert.True(t, channel.EqualWireMaps(recvAddr0, wire.AddressMapfromAccountMap(account0)))
	}()

	err := ExchangeAddrsActive(context.Background(), account0, wire.AddressMapfromAccountMap(account1), conn0)
	assert.NoError(t, err)

	wg.Wait()
}

func TestExchangeAddrs_Timeout(t *testing.T) {
	rng := test.Prng(t)
	a, _ := newPipeConnPair()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ctxtest.AssertTerminates(t, 2*timeout, func() {
		addr, err := ExchangeAddrsPassive(ctx, wiretest.NewRandomAccountMap(rng, channel.TestBackendID), a)
		assert.Nil(t, addr)
		assert.Error(t, err)
	})
}

func TestExchangeAddrs_BogusMsg(t *testing.T) {
	rng := test.Prng(t)
	acc := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	conn := newMockConn()
	conn.recvQueue <- wiretest.NewRandomEnvelope(rng, wire.NewPingMsg())
	addr, err := ExchangeAddrsPassive(context.Background(), acc, conn)

	assert.Error(t, err, "ExchangeAddrs should error when peer sends a non-AuthResponseMsg")
	assert.Nil(t, addr)
}
