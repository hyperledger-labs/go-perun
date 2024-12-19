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
	"crypto/tls"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	perunio "perun.network/go-perun/wire/perunio/serializer"

	"polycry.pt/poly-go/context/test"
)

const addr = "0.0.0.0:1337"

// serverKey and serverCert are generated with the following commands:
// openssl ecparam -genkey -name prime256v1 -out server.key
// openssl req -new -x509 -key server.key -out server.pem -days 3650.
const testServerKey = `-----BEGIN EC PARAMETERS-----
BggqhkjOPQMBBw==
-----END EC PARAMETERS-----
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIHg+g2unjA5BkDtXSN9ShN7kbPlbCcqcYdDu+QeV8XWuoAoGCCqGSM49
AwEHoUQDQgAEcZpodWh3SEs5Hh3rrEiu1LZOYSaNIWO34MgRxvqwz1FMpLxNlx0G
cSqrxhPubawptX5MSr02ft32kfOlYbaF5Q==
-----END EC PRIVATE KEY-----
`

const testServerCert = `-----BEGIN CERTIFICATE-----
MIIB+TCCAZ+gAwIBAgIJAL05LKXo6PrrMAoGCCqGSM49BAMCMFkxCzAJBgNVBAYT
AkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRn
aXRzIFB0eSBMdGQxEjAQBgNVBAMMCWxvY2FsaG9zdDAeFw0xNTEyMDgxNDAxMTNa
Fw0yNTEyMDUxNDAxMTNaMFkxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0
YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxEjAQBgNVBAMM
CWxvY2FsaG9zdDBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABHGaaHVod0hLOR4d
66xIrtS2TmEmjSFjt+DIEcb6sM9RTKS8TZcdBnEqq8YT7m2sKbV+TEq9Nn7d9pHz
pWG2heWjUDBOMB0GA1UdDgQWBBR0fqrecDJ44D/fiYJiOeBzfoqEijAfBgNVHSME
GDAWgBR0fqrecDJ44D/fiYJiOeBzfoqEijAMBgNVHRMEBTADAQH/MAoGCCqGSM49
BAMCA0gAMEUCIEKzVMF3JqjQjuM2rX7Rx8hancI5KJhwfeKu1xbyR7XaAiEA2UT7
1xOP035EcraRmWPe7tO0LpXgMxlh2VItpc2uc2w=
-----END CERTIFICATE-----
`

func TestNewTCPListener(t *testing.T) {
	cer, err := tls.X509KeyPair([]byte(testServerCert), []byte(testServerKey))
	require.NoError(t, err, "loading server key and cert")
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12, // Set minimum TLS version to TLS 1.2
		Certificates: []tls.Certificate{cer},
	}
	l, err := NewTCPListener(addr, tlsConfig)
	require.NoError(t, err)
	defer l.Close()
}

func TestNewUnixListener(t *testing.T) {
	cer, err := tls.X509KeyPair([]byte(testServerCert), []byte(testServerKey))
	require.NoError(t, err, "loading server key and cert")
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12, // Set minimum TLS version to TLS 1.2
		Certificates: []tls.Certificate{cer},
	}
	l, err := NewUnixListener(addr, tlsConfig)
	require.NoError(t, err)
	defer l.Close()
}

func TestListener_Close(t *testing.T) {
	cer, err := tls.X509KeyPair([]byte(testServerCert), []byte(testServerKey))
	require.NoError(t, err, "loading server key and cert")
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12, // Set minimum TLS version to TLS 1.2
		Certificates: []tls.Certificate{cer},
	}
	t.Run("double close", func(t *testing.T) {
		l, err := NewTCPListener(addr, tlsConfig)
		require.NoError(t, err)
		assert.NoError(t, l.Close(), "first close must not return error")
		assert.Error(t, l.Close(), "second close must result in error")
	})
}

func TestNewListener(t *testing.T) {
	cer, err := tls.X509KeyPair([]byte(testServerCert), []byte(testServerKey))
	require.NoError(t, err, "loading server key and cert")
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12, // Set minimum TLS version to TLS 1.2
		Certificates: []tls.Certificate{cer},
	}
	t.Run("happy", func(t *testing.T) {
		l, err := NewTCPListener(addr, tlsConfig)
		assert.NoError(t, err)
		require.NotNil(t, l)
		l.Close()
	})

	t.Run("sad", func(t *testing.T) {
		l, err := NewTCPListener("not an address", tlsConfig)
		assert.Error(t, err)
		assert.Nil(t, l)
	})

	t.Run("address in use", func(t *testing.T) {
		l, err := NewTCPListener(addr, tlsConfig)
		require.NoError(t, err)
		_, err = NewTCPListener(addr, tlsConfig)
		require.Error(t, err)
		l.Close()
	})
}

func TestListener_Accept(t *testing.T) {
	cer, err := tls.X509KeyPair([]byte(testServerCert), []byte(testServerKey))
	require.NoError(t, err, "loading server key and cert")
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12, // Set minimum TLS version to TLS 1.2
		Certificates: []tls.Certificate{cer},
	}
	// Happy case already tested in TestDialer_Dial.
	ser := perunio.Serializer()
	timeout := 100 * time.Millisecond
	t.Run("timeout", func(t *testing.T) {
		l, err := NewTCPListener(addr, tlsConfig)
		require.NoError(t, err)
		defer l.Close()

		test.AssertNotTerminates(t, timeout, func() {
			l.Accept(ser) //nolint:errcheck
		})
	})

	t.Run("closed", func(t *testing.T) {
		l, err := NewTCPListener(addr, tlsConfig)
		require.NoError(t, err)
		l.Close()

		test.AssertTerminates(t, timeout, func() {
			conn, err := l.Accept(ser)
			assert.Nil(t, conn)
			assert.Error(t, err)
		})
	})
}
