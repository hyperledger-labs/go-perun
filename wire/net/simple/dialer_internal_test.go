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
	lConfig, dConfig, err := generateSelfSignedCertConfigs(commonName, sans)
	require.NoError(t, err, "failed to generate self-signed certificate configs")

	l, err := NewTCPListener(lhost, lConfig)
	require.NoError(t, err)
	defer l.Close()

	ser := perunio.Serializer()
	d := NewTCPDialer(timeout, dConfig)
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

// generateSelfSignedCertConfigs generates a self-signed certificate and returns
// the server and client TLS configurations.
func generateSelfSignedCertConfigs(commonName string, sans []string) (*tls.Config, *tls.Config, error) {
	keySize := 2048
	// Generate a new RSA private key for the server
	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}

	// Generate a new RSA private key for the client
	clientPrivateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}

	// Create a certificate template for the server
	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Perun Network"},
			CommonName:   commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add SANs to the server certificate template
	for _, san := range sans {
		if ip := net.ParseIP(san); ip != nil {
			serverTemplate.IPAddresses = append(serverTemplate.IPAddresses, ip)
		} else {
			serverTemplate.DNSNames = append(serverTemplate.DNSNames, san)
		}
	}

	// Generate a self-signed server certificate
	serverCertDER, err := x509.CreateCertificate(rand.Reader, &serverTemplate, &serverTemplate, &serverPrivateKey.PublicKey, serverPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	// Encode the server certificate to PEM format
	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})

	// Encode the server private key to PEM format
	serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey)})

	// Create a tls.Certificate object for the server
	serverCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		return nil, nil, err
	}

	// Create a certificate template for the client
	clientTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Perun Network"},
			CommonName:   commonName, // Change this to the client's common name
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, // Set the client authentication usage
		BasicConstraintsValid: true,
	}

	// Generate a self-signed client certificate
	clientCertDER, err := x509.CreateCertificate(rand.Reader, &clientTemplate, &clientTemplate, &clientPrivateKey.PublicKey, serverPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	// Encode the client certificate to PEM format
	clientCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertDER})

	// Encode the client private key to PEM format
	clientKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientPrivateKey)})

	// Create a tls.Certificate object for the client
	clientCert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
	if err != nil {
		return nil, nil, err
	}

	serverCertPool := x509.NewCertPool()
	ok := serverCertPool.AppendCertsFromPEM(clientCertPEM)
	if !ok {
		return nil, nil, errors.New("failed to parse root certificate")
	}

	// Create the server-side TLS configuration
	serverConfig := &tls.Config{
		ClientCAs:    serverCertPool,
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12, // Set minimum TLS version to TLS 1.2
	}

	clientCertPool := x509.NewCertPool()
	ok = clientCertPool.AppendCertsFromPEM(serverCertPEM)
	if !ok {
		return nil, nil, errors.New("failed to parse root certificate")
	}

	// Create the client-side TLS configuration
	clientConfig := &tls.Config{
		RootCAs:      clientCertPool,
		Certificates: []tls.Certificate{clientCert},
		MinVersion:   tls.VersionTLS12, // Set minimum TLS version to TLS 1.2
	}

	return serverConfig, clientConfig, nil
}
