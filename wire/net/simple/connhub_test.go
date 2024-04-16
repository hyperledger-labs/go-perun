// Copyright 2019 - See NOTICE file for copyright holders.
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
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "perun.network/go-perun/backend/sim" // backend init
	"perun.network/go-perun/wire"
	perunio "perun.network/go-perun/wire/perunio/serializer"
	wiretest "perun.network/go-perun/wire/test"
	ctxtest "polycry.pt/poly-go/context/test"
	"polycry.pt/poly-go/sync"
	pkgtest "polycry.pt/poly-go/test"
)

var (
	commonName = "127.0.0.1"
)

var sans = []string{"127.0.0.1", "localhost"}

func TestConnHub_Create(t *testing.T) {
	rng := pkgtest.Prng(t)
	ser := perunio.Serializer()

	t.Run("create and dial existing", func(t *testing.T) {
		assert := assert.New(t)

		hosts := make([]string, 2)
		for i := 0; i < 2; i++ {
			port, err := findFreePort()
			assert.NoError(err, "finding free port")
			hosts[i] = fmt.Sprintf("127.0.0.1:%d", port)
		}
		tlsConfigs, err := GenerateSelfSignedCertConfigs(commonName, sans, 2)
		assert.NoError(err, "generating self-signed cert configs")

		var c ConnHub
		addr := NewRandomAddress(rng)
		d, l := c.NewNetDialer(DefaultTimeout, tlsConfigs[0]), c.NewNetListener(addr, hosts[0], tlsConfigs[0])
		assert.NotNil(d)
		assert.NotNil(l)

		ct := pkgtest.NewConcurrent(t)
		go ctxtest.AssertTerminates(t, timeout, func() {
			ct.Stage("accept", func(rt pkgtest.ConcT) {
				conn, err := l.Accept(ser)
				assert.NoError(err)
				require.NotNil(rt, conn)
				assert.NoError(conn.Send(wiretest.NewRandomEnvelope(rng, wire.NewPingMsg())))
			})
		})

		ctxtest.AssertTerminates(t, timeout, func() {
			ct.Stage("dial", func(rt pkgtest.ConcT) {
				conn, err := d.Dial(context.Background(), addr, ser)
				assert.NoError(err)
				require.NotNil(rt, conn)
				m, err := conn.Recv()
				assert.NoError(err)
				assert.IsType(wire.NewPingMsg(), m.Msg)
			})
		})

		ct.Wait("accept", "dial")
	})

	t.Run("double create", func(t *testing.T) {
		assert := assert.New(t)

		hosts := make([]string, 2)
		for i := 0; i < 2; i++ {
			port, err := findFreePort()
			assert.NoError(err, "finding free port")
			hosts[i] = fmt.Sprintf("127.0.0.1:%d", port)
		}
		tlsConfigs, err := GenerateSelfSignedCertConfigs(commonName, sans, 2)
		assert.NoError(err, "generating self-signed cert configs")

		var c ConnHub
		addr := NewRandomAddress(rng)

		l := c.NewNetListener(addr, hosts[0], tlsConfigs[1])
		assert.NotNil(l)

		assert.Panics(func() { c.NewNetListener(addr, hosts[1], tlsConfigs[1]) })
	})

	t.Run("dial nonexisting", func(t *testing.T) {
		assert := assert.New(t)

		hosts := make([]string, 2)
		for i := 0; i < 2; i++ {
			port, err := findFreePort()
			assert.NoError(err, "finding free port")
			hosts[i] = fmt.Sprintf("127.0.0.1:%d", port)
		}
		tlsConfigs, err := GenerateSelfSignedCertConfigs(commonName, sans, 2)
		assert.NoError(err, "generating self-signed cert configs")

		var c ConnHub

		d := c.NewNetDialer(DefaultTimeout, tlsConfigs[0])
		ctxtest.AssertTerminates(t, timeout, func() {
			conn, err := d.Dial(context.Background(), NewRandomAddress(rng), ser)
			assert.Nil(conn)
			assert.Error(err)
		})
	})

	t.Run("closed create", func(t *testing.T) {
		assert := assert.New(t)

		hosts := make([]string, 2)
		for i := 0; i < 2; i++ {
			port, err := findFreePort()
			assert.NoError(err, "finding free port")
			hosts[i] = fmt.Sprintf("127.0.0.1:%d", port)
		}
		tlsConfigs, err := GenerateSelfSignedCertConfigs(commonName, sans, 2)
		assert.NoError(err, "generating self-signed cert configs")

		var c ConnHub
		c.Close()
		addr := NewRandomAddress(rng)

		assert.Panics(func() { c.NewNetDialer(DefaultTimeout, tlsConfigs[0]) })
		assert.Panics(func() { c.NewNetListener(addr, hosts[0], tlsConfigs[0]) })
	})
}

func TestConnHub_Close(t *testing.T) {
	rng := pkgtest.Prng(t)
	t.Run("nonempty close", func(t *testing.T) {
		assert := assert.New(t)

		hosts := make([]string, 2)
		for i := 0; i < 2; i++ {
			port, err := findFreePort()
			assert.NoError(err, "finding free port")
			hosts[i] = fmt.Sprintf("127.0.0.1:%d", port)
		}
		tlsConfigs, err := GenerateSelfSignedCertConfigs(commonName, sans, 2)
		assert.NoError(err, "generating self-signed cert configs")

		var c ConnHub
		l := c.NewNetListener(NewRandomAddress(rng), hosts[0], tlsConfigs[0])
		assert.NoError(c.Close())
		assert.True(l.IsClosed())
	})

	t.Run("nonempty close with error (listener)", func(t *testing.T) {
		assert := assert.New(t)

		hosts := make([]string, 2)
		for i := 0; i < 2; i++ {
			port, err := findFreePort()
			assert.NoError(err, "finding free port")
			hosts[i] = fmt.Sprintf("127.0.0.1:%d", port)
		}
		tlsConfigs, err := GenerateSelfSignedCertConfigs(commonName, sans, 2)
		assert.NoError(err, "generating self-signed cert configs")

		var c ConnHub
		l := c.NewNetListener(NewRandomAddress(rng), hosts[0], tlsConfigs[0])
		l2, err := NewTCPListener(hosts[1], tlsConfigs[1])
		assert.NoError(err, "creating listener")
		l2.Close()
		err = c.insertListener(NewRandomAccount(rng).Address(), l2)
		assert.NoError(err)
		assert.Error(c.Close())
		assert.True(l.IsClosed())
	})

	t.Run("nonempty close with error (dialer)", func(t *testing.T) {
		assert := assert.New(t)

		hosts := make([]string, 2)
		for i := 0; i < 2; i++ {
			port, err := findFreePort()
			assert.NoError(err, "finding free port")
			hosts[i] = fmt.Sprintf("127.0.0.1:%d", port)
		}
		tlsConfigs, err := GenerateSelfSignedCertConfigs(commonName, sans, 2)
		assert.NoError(err, "generating self-signed cert configs")

		var c ConnHub
		d := c.NewNetDialer(DefaultTimeout, tlsConfigs[0])
		d2 := &Dialer{}
		d2.Close()
		c.insertDialer(d2)
		assert.Error(c.Close())
		assert.True(d.IsClosed())
	})

	t.Run("double close", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		assert.NoError(c.Close())
		err := c.Close()
		assert.Error(err)
		assert.True(sync.IsAlreadyClosedError(err))
	})
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
