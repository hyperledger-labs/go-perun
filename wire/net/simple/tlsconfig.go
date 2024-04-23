// Copyright 2022 - See NOTICE file for copyright holders.
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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// GenerateSelfSignedCertConfigs generates self-signed certificates and returns
// a list of TLS configurations for n clients.
func GenerateSelfSignedCertConfigs(commonName string, sans []string, numClients int) ([]*tls.Config, error) {
	keySize := 2048
	configs := make([]*tls.Config, numClients)
	certPEMs := make([][]byte, numClients)
	tlsCerts := make([]tls.Certificate, numClients)

	for i := 0; i < numClients; i++ {
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
			NotAfter:              time.Now().Add(24 * time.Hour),
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

	for i := 0; i < numClients; i++ {
		certPool := x509.NewCertPool()
		for j := 0; j < numClients; j++ {
			ok := certPool.AppendCertsFromPEM(certPEMs[j])
			if !ok {
				return nil, fmt.Errorf("failed to parse root certificate")
			}
		}

		// Create the server-side TLS configuration
		configs[i] = &tls.Config{
			RootCAs:      certPool,
			ClientCAs:    certPool,
			Certificates: []tls.Certificate{tlsCerts[i]},

			MinVersion: tls.VersionTLS12, // Set minimum TLS version to TLS 1.2
		}
	}

	return configs, nil
}
