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

package simple

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	simwallet "perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	_ "perun.network/go-perun/wire/protobuf" // wire serialzer init
	ctxtest "polycry.pt/poly-go/context/test"
	"polycry.pt/poly-go/test"
)

func TestNewTCPDialer(t *testing.T) {
	d := NewTCPDialer(0)
	assert.Equal(t, d.network, "tcp")
}

func TestNewUnixDialer(t *testing.T) {
	d := NewUnixDialer(0)
	assert.Equal(t, d.network, "unix")
}

func TestDialer_Register(t *testing.T) {
	rng := test.Prng(t)
	addr := simwallet.NewRandomAddress(rng)
	key := wallet.Key(addr)
	d := NewTCPDialer(0)

	_, ok := d.host(key)
	require.False(t, ok)

	d.Register(addr, "host")

	host, ok := d.host(key)
	assert.True(t, ok)
	assert.Equal(t, host, "host")
}

func TestDialer_Dial(t *testing.T) {
	timeout := 100 * time.Millisecond
	rng := test.Prng(t)
	lhost := "127.0.0.1:7357"
	laddr := simwallet.NewRandomAddress(rng)

	l, err := NewTCPListener(lhost)
	require.NoError(t, err)
	defer l.Close()

	d := NewTCPDialer(timeout)
	d.Register(laddr, lhost)
	daddr := simwallet.NewRandomAddress(rng)
	defer d.Close()

	t.Run("happy", func(t *testing.T) {
		e := &wire.Envelope{
			Sender:    daddr,
			Recipient: laddr,
			Msg:       wire.NewPingMsg(),
		}
		ct := test.NewConcurrent(t)
		go ct.Stage("accept", func(rt test.ConcT) {
			conn, err := l.Accept()
			assert.NoError(t, err)
			require.NotNil(rt, conn)

			re, err := conn.Recv()
			assert.NoError(t, err)
			assert.Equal(t, re, e)
		})

		ct.Stage("dial", func(rt test.ConcT) {
			ctxtest.AssertTerminates(t, timeout, func() {
				conn, err := d.Dial(context.Background(), laddr)
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
			conn, err := d.Dial(ctx, laddr)
			assert.Nil(t, conn)
			assert.Error(t, err)
		})
	})

	t.Run("unknown host", func(t *testing.T) {
		noHostAddr := simwallet.NewRandomAddress(rng)
		d.Register(noHostAddr, "no such host")

		ctxtest.AssertTerminates(t, timeout, func() {
			conn, err := d.Dial(context.Background(), noHostAddr)
			assert.Nil(t, conn)
			assert.Error(t, err)
		})
	})

	t.Run("unknown address", func(t *testing.T) {
		ctxtest.AssertTerminates(t, timeout, func() {
			unkownAddr := simwallet.NewRandomAddress(rng)
			conn, err := d.Dial(context.Background(), unkownAddr)
			assert.Error(t, err)
			assert.Nil(t, conn)
		})
	})
}
