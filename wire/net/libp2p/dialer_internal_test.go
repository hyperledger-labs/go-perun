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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	perunio "perun.network/go-perun/wire/perunio/serializer"
	"perun.network/go-perun/wire/test"

	ctxtest "polycry.pt/poly-go/context/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestNewDialer(t *testing.T) {
	rng := pkgtest.Prng(t)
	h := getHost(rng)

	d := NewP2PDialer(h)
	assert.NotNil(t, d)
	d.Close()
}

func TestDialer_Register(t *testing.T) {
	rng := pkgtest.Prng(t)
	addr := NewRandomAddress(rng)
	key := wire.Key(addr)
	h := getHost(rng)
	d := NewP2PDialer(h)
	defer d.Close()

	_, ok := d.get(key)
	require.False(t, ok)

	addrs := make(map[wallet.BackendID]wire.Address)
	addrs[test.TestBackendID] = addr
	d.Register(addrs, "p2pAddress")

	host, ok := d.get(key)
	assert.True(t, ok)
	assert.Equal(t, host, "p2pAddress")
}

func TestDialer_Dial(t *testing.T) {
	timeout := 2 * time.Second
	rng := pkgtest.Prng(t)

	lHost := getHost(rng)
	lAddr := lHost.Address()
	laddrs := make(map[wallet.BackendID]wire.Address)
	laddrs[test.TestBackendID] = lAddr
	lpeerID := lHost.ID()
	listener := NewP2PListener(lHost)
	defer listener.Close()

	dHost := getHost(rng)
	dAddr := dHost.Address()
	daddrs := make(map[wallet.BackendID]wire.Address)
	daddrs[test.TestBackendID] = dAddr
	dialer := NewP2PDialer(dHost)
	dialer.Register(laddrs, lpeerID.String())
	defer dialer.Close()

	t.Run("happy", func(t *testing.T) {
		e := &wire.Envelope{
			Sender:    daddrs,
			Recipient: laddrs,
			Msg:       wire.NewPingMsg(),
		}
		ct := pkgtest.NewConcurrent(t)

		go ct.Stage("accept", func(rt pkgtest.ConcT) {
			conn, err := listener.Accept(perunio.Serializer())
			assert.NoError(t, err)
			require.NotNil(rt, conn)

			re, err := conn.Recv()
			assert.NoError(t, err)
			assert.Equal(t, re, e)
		})

		ct.Stage("dial", func(rt pkgtest.ConcT) {
			ctxtest.AssertTerminates(t, timeout, func() {
				conn, err := dialer.Dial(context.Background(), laddrs, perunio.Serializer())
				assert.NoError(t, err)
				require.NotNil(rt, conn)

				assert.NoError(t, conn.Send(e))
			})
		})

		ct.Wait("dial", "accept")
	})

	t.Run("aborted context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		ctxtest.AssertTerminates(t, timeout, func() {
			conn, err := dialer.Dial(ctx, laddrs, perunio.Serializer())
			assert.Nil(t, conn)
			assert.Error(t, err)
		})
	})

	t.Run("unknown host", func(t *testing.T) {
		noHostAddr := map[wallet.BackendID]wire.Address{test.TestBackendID: NewRandomAddress(rng)}
		dialer.Register(noHostAddr, "no such host")

		ctxtest.AssertTerminates(t, timeout, func() {
			conn, err := dialer.Dial(context.Background(), noHostAddr, perunio.Serializer())
			assert.Nil(t, conn)
			assert.Error(t, err)
		})
	})

	t.Run("unknown address", func(t *testing.T) {
		ctxtest.AssertTerminates(t, timeout, func() {
			unknownAddr := map[wallet.BackendID]wire.Address{test.TestBackendID: NewRandomAddress(rng)}
			conn, err := dialer.Dial(context.Background(), unknownAddr, perunio.Serializer())
			assert.Error(t, err)
			assert.Nil(t, conn)
		})
	})
}

func TestDialer_Close(t *testing.T) {
	t.Run("double close", func(t *testing.T) {
		rng := pkgtest.Prng(t)
		h := getHost(rng)
		d := NewP2PDialer(h)

		assert.NoError(t, d.Close(), "first close must not return error")
		assert.Error(t, d.Close(), "second close must result in error")
	})
}
