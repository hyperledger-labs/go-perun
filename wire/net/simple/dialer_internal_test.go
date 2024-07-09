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
	"crypto/tls"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/wire"
	perunio "perun.network/go-perun/wire/perunio/serializer"
	wiretest "perun.network/go-perun/wire/test"
	ctxtest "polycry.pt/poly-go/context/test"
	"polycry.pt/poly-go/test"
)

func TestNewTCPDialer(t *testing.T) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Set minimum TLS version to TLS 1.2
	}
	d := NewTCPDialer(0, tlsConfig)
	assert.Equal(t, d.network, "tcp")
}

func TestNewUnixDialer(t *testing.T) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Set minimum TLS version to TLS 1.2
	}
	d := NewUnixDialer(0, tlsConfig)
	assert.Equal(t, d.network, "unix")
}

func TestDialer_Register(t *testing.T) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Set minimum TLS version to TLS 1.2
	}
	rng := test.Prng(t)
	addr := NewRandomAddress(rng)
	key := wire.Key(addr)
	d := NewTCPDialer(0, tlsConfig)

	_, ok := d.host(key)
	require.False(t, ok)

	d.Register(addr, "host")

	host, ok := d.host(key)
	assert.True(t, ok)
	assert.Equal(t, host, "host")
}
func TestDialer_Dial(t *testing.T) {
	timeout := 10000 * time.Millisecond
	rng := test.Prng(t)
	lhost := "127.0.0.1:7355"
	laddr := wiretest.NewRandomAccount(rng).Address()

	commonName := "127.0.0.1"
	sans := []string{"127.0.0.1", "localhost"}
	configs, err := GenerateSelfSignedCertConfigs(commonName, sans, 2)
	require.NoError(t, err, "failed to generate self-signed certificate configs")

	l, err := NewTCPListener(lhost, configs[0])
	require.NoError(t, err)
	defer l.Close()

	ser := perunio.Serializer()
	d := NewTCPDialer(timeout, configs[1])
	d.Register(laddr, lhost)
	daddr := wiretest.NewRandomAccount(rng).Address()
	defer d.Close()

	t.Run("happy", func(t *testing.T) {
		e := &wire.Envelope{
			Sender:    daddr,
			Recipient: laddr,
			Msg:       wire.NewPingMsg(),
		}
		ct := test.NewConcurrent(t)
		go ct.Stage("accept", func(rt test.ConcT) {
			conn, err := l.Accept(ser)
			assert.NoError(t, err)
			require.NotNil(rt, conn)

			re, err := conn.Recv()
			assert.NoError(t, err)
			assert.Equal(t, re, e)
		})

		ct.Stage("dial", func(rt test.ConcT) {
			ctxtest.AssertTerminates(t, timeout, func() {
				conn, err := d.Dial(context.Background(), laddr, ser)
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
			conn, err := d.Dial(ctx, laddr, ser)
			assert.Nil(t, conn)
			assert.Error(t, err)
		})
	})

	t.Run("unknown host", func(t *testing.T) {
		noHostAddr := NewRandomAddress(rng)

		ctxtest.AssertTerminates(t, timeout, func() {
			conn, err := d.Dial(context.Background(), noHostAddr, ser)
			assert.Nil(t, conn)
			assert.Error(t, err)
		})
	})

	t.Run("unknown address", func(t *testing.T) {
		ctxtest.AssertTerminates(t, timeout, func() {
			unkownAddr := NewRandomAddress(rng)
			conn, err := d.Dial(context.Background(), unkownAddr, ser)
			assert.Error(t, err)
			assert.Nil(t, conn)
		})
	})
}
