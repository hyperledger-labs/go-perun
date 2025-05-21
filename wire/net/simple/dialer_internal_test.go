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
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"

	"perun.network/go-perun/channel"

	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"
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
	assert.Equal(t, "tcp", d.network)
}

func TestNewUnixDialer(t *testing.T) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Set minimum TLS version to TLS 1.2
	}
	d := NewUnixDialer(0, tlsConfig)
	assert.Equal(t, "unix", d.network)
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

	d.Register(map[wallet.BackendID]wire.Address{channel.TestBackendID: addr}, "host")

	host, ok := d.host(key)
	assert.True(t, ok)
	assert.Equal(t, "host", host)
}

func TestDialer_Dial(t *testing.T) {
	timeout := 100 * time.Millisecond
	rng := test.Prng(t)
	lhost := "127.0.0.1:7357"
	laddr := wire.AddressMapfromAccountMap(wiretest.NewRandomAccountMap(rng, channel.TestBackendID))

	commonName := "127.0.0.1"
	sans := []string{"127.0.0.1", "localhost"}
	tlsConfigs, err := generateSelfSignedCertConfigs(commonName, sans, 2)
	require.NoError(t, err, "failed to generate self-signed certificate configs")

	l, err := NewTCPListener(lhost, tlsConfigs[0])
	require.NoError(t, err)
	defer l.Close()

	ser := perunio.Serializer()
	d := NewTCPDialer(timeout, tlsConfigs[1])
	d.Register(laddr, lhost)
	daddr := wire.AddressMapfromAccountMap(wiretest.NewRandomAccountMap(rng, channel.TestBackendID))
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
			assert.NotNil(rt, conn)

			re, err := conn.Recv()
			require.NoError(t, err)
			assert.Equal(t, re, e)
		})

		ct.Stage("dial", func(rt test.ConcT) {
			ctxtest.AssertTerminates(t, timeout, func() {
				conn, err := d.Dial(context.Background(), laddr, ser)
				require.NoError(t, err)
				require.NotNil(rt, conn)

				require.NoError(t, conn.Send(e))
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
		noHostAddr := NewRandomAddresses(rng, []wallet.BackendID{wiretest.TestBackendID})
		d.Register(noHostAddr, "no such host")

		ctxtest.AssertTerminates(t, timeout, func() {
			conn, err := d.Dial(context.Background(), noHostAddr, ser)
			assert.Nil(t, conn)
			assert.Error(t, err)
		})
	})

	t.Run("unknown address", func(t *testing.T) {
		ctxtest.AssertTerminates(t, timeout, func() {
			unkownAddr := NewRandomAddresses(rng, []wallet.BackendID{wiretest.TestBackendID})
			conn, err := d.Dial(context.Background(), unkownAddr, ser)
			assert.Error(t, err)
			assert.Nil(t, conn)
		})
	})
}

const certificateTimeout = time.Hour

// generateSelfSignedCertConfigs generates self-signed certificates and returns
// a list of TLS configurations for n clients.
func generateSelfSignedCertConfigs(commonName string, sans []string, numClients int) ([]*tls.Config, error) {
	keySize := 2048
	configs := make([]*tls.Config, numClients)
	certPEMs := make([][]byte, numClients)
	tlsCerts := make([]tls.Certificate, numClients)

	for i := range numClients {
		// Private key for the client
		privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
		if err != nil {
			return nil, err
		}

		// Create a certificate template
		template := x509.Certificate{
			SerialNumber: big.NewInt(int64(i) + 1),
			Subject: pkix.Name{
				Organization: []string{"Perun Network"},
				CommonName:   fmt.Sprintf("%s-client-%d", commonName, i+1),
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(certificateTimeout),
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
			BasicConstraintsValid: true,
		}

		// Add SANs to the server certificate template
		for _, san := range sans {
			if ip := net.ParseIP(san); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			} else {
				template.DNSNames = append(template.DNSNames, san)
			}
		}

		// Generate a self-signed server certificate
		certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
		if err != nil {
			return nil, err
		}

		// Encode the server certificate to PEM format
		certPEMs[i] = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

		// Encode the server private key to PEM format
		keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

		// Create a tls.Certificate object for the server
		tlsCerts[i], err = tls.X509KeyPair(certPEMs[i], keyPEM)
		if err != nil {
			return nil, err
		}
	}

	for i := range numClients {
		certPool := x509.NewCertPool()
		for j := range numClients {
			ok := certPool.AppendCertsFromPEM(certPEMs[j])
			if !ok {
				return nil, errors.New("failed to parse root certificate")
			}
		}

		// Create the server-side TLS configuration
		configs[i] = &tls.Config{
			RootCAs:      certPool,
			ClientCAs:    certPool,
			Certificates: []tls.Certificate{tlsCerts[i]},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			MinVersion:   tls.VersionTLS12, // Set minimum TLS version to TLS 1.2
		}
	}

	return configs, nil
}
