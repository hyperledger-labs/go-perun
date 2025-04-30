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
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	swarm "github.com/libp2p/go-libp2p/p2p/net/swarm"
	libp2pclient "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

const (
	relayID = "QmVCPfUMr98PaaM8qbAQBgJ9jqc7XHpGp7AsyragdFDmgm"

	keySize = 2048

	queryProtocol    = "/address-book/query/1.0.0"    // Protocol for querying the relay-server for a peerID.
	registerProtocol = "/address-book/register/1.0.0" // Protocol for registering an on-chain address with the relay-server.
	removeProtocol   = "/address-book/remove/1.0.0"   // Protocol for deregistering an on-chain address with the relay-server.
)

// Account represents a libp2p wire account containing a libp2p host.
type Account struct {
	host.Host
	relayAddr   string
	privateKey  crypto.PrivKey
	reservation *libp2pclient.Reservation
	closer      context.CancelFunc
}

// NewRandomAccount generates a new random account.
func NewRandomAccount(rng *rand.Rand) *Account {
	relayInfo, relayAddr, err := getRelayServerInfo()
	if err != nil {
		panic(err)
	}

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, keySize, rng)
	if err != nil {
		panic(err)
	}

	// Construct a new libp2p client for our relay-server.
	// Identity(prvKey)		- Use a RSA private key to generate the ID of the host.
	// EnableRelay()		- Enable relay system and configures itself as a node behind a NAT
	client, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.Identity(prvKey),
		libp2p.EnableRelay(),
	)
	if err != nil {
		client.Close()
		panic(err)
	}

	// Redialing hacked
	if sw, ok := client.Network().(*swarm.Swarm); ok {
		sw.Backoff().Clear(relayInfo.ID)
	}

	if err := client.Connect(context.Background(), *relayInfo); err != nil {
		client.Close()
		panic(errors.WithMessage(err, "connecting to the relay server"))
	}

	// Reserve connection
	// Hosts that want to have messages relayed on their behalf need to reserve a slot
	// with the circuit relay service host
	resv, err := libp2pclient.Reserve(context.Background(), client, *relayInfo)
	if err != nil {
		panic(errors.WithMessage(err, "failed to receive a relay reservation from relay server"))
	}

	ctx, cancel := context.WithCancel(context.Background())
	acc := &Account{client, relayAddr, prvKey, resv, cancel}

	go acc.keepReservationAlive(ctx, *relayInfo)

	return acc
}

// NewAccountFromPrivateKeyBytes creates a new account from a given private key.
func NewAccountFromPrivateKeyBytes(prvKeyBytes []byte) (*Account, error) {
	prvKey, err := crypto.UnmarshalPrivateKey(prvKeyBytes)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshalling private key")
	}

	relayInfo, relayAddr, err := getRelayServerInfo()
	if err != nil {
		panic(err)
	}
	// Construct a new libp2p client for our relay-server.
	// Identity(prvKey)		- Use a RSA private key to generate the ID of the host.
	// EnableRelay()		- Enable relay system and configures itself as a node behind a NAT
	client, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.Identity(prvKey),
		libp2p.EnableRelay(),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "creating new libp2p client")
	}

	if sw, ok := client.Network().(*swarm.Swarm); ok {
		sw.Backoff().Clear(relayInfo.ID)
	}

	if err := client.Connect(context.Background(), *relayInfo); err != nil {
		client.Close()
		return nil, errors.WithMessage(err, "connecting to the relay server")
	}

	// Reserve connection
	// Hosts that want to have messages relayed on their behalf need to reserve a slot
	// with the circuit relay service host
	res, err := libp2pclient.Reserve(context.Background(), client, *relayInfo)
	if err != nil {
		panic(errors.WithMessage(err, "failed to receive a relay reservation from relay server"))
	}

	ctx, cancel := context.WithCancel(context.Background())
	acc := &Account{client, relayAddr, prvKey, res, cancel}

	go acc.keepReservationAlive(ctx, *relayInfo)

	return acc, nil
}

// Address returns the account's address.
func (acc *Account) Address() wire.Address {
	return &Address{acc.ID()}
}

// Sign signs the given message with the account's private key.
func (acc *Account) Sign(data []byte) ([]byte, error) {
	// Extract the private key from the account.
	if acc.privateKey == nil {
		return nil, errors.New("private key not set")
	}
	hashed := sha256.Sum256(data)

	signature, err := acc.privateKey.Sign(hashed[:])
	if err != nil {
		return nil, err
	}
	return signature, nil
}

// MarshalPrivateKey marshals the account's private key to binary.
func (acc *Account) MarshalPrivateKey() ([]byte, error) {
	return crypto.MarshalPrivateKey(acc.privateKey)
}

// RegisterOnChainAddress registers an on-chain address with the account to the relay-server's address book.
func (acc *Account) RegisterOnChainAddress(onChainAddr wallet.Address) error {
	id, err := peer.Decode(relayID)
	if err != nil {
		err = errors.WithMessage(err, "decoding peer id of relay server")
		return err
	}

	s, err := acc.NewStream(network.WithAllowLimitedConn(context.Background(), registerProtocol[1:]), id, registerProtocol)
	if err != nil {
		return errors.WithMessage(err, "creating new stream")
	}
	defer s.Close()

	var registerData struct {
		OnChainAddress string
		PeerID         string
	}
	if onChainAddr == nil {
		return errors.New("on-chain address is nil")
	}
	registerData.OnChainAddress = onChainAddr.String()
	registerData.PeerID = acc.ID().String()

	data, err := json.Marshal(registerData) //nolint:musttag
	if err != nil {
		return errors.WithMessage(err, "marshalling register data")
	}

	_, err = s.Write(data)
	if err != nil {
		return errors.WithMessage(err, "writing register data")
	}

	return nil
}

// Close closes the account.
func (acc *Account) Close() error {
	acc.closer()
	return acc.Host.Close()
}

// DeregisterOnChainAddress deregisters an on-chain address with the account from the relay-server's address book.
func (acc *Account) DeregisterOnChainAddress(onChainAddr wallet.Address) error {
	relayInfo, _, err := getRelayServerInfo()
	if err != nil {
		return errors.WithMessage(err, "getting relay server info")
	}

	s, err := acc.NewStream(network.WithAllowLimitedConn(context.Background(), removeProtocol[1:]), relayInfo.ID, removeProtocol)
	if err != nil {
		return errors.WithMessage(err, "creating new stream")
	}
	defer s.Close()

	var unregisterData struct {
		OnChainAddress string
		PeerID         string
	}
	unregisterData.OnChainAddress = onChainAddr.String()
	unregisterData.PeerID = acc.ID().String()

	data, err := json.Marshal(unregisterData) //nolint:musttag
	if err != nil {
		return errors.WithMessage(err, "marshalling register data")
	}

	_, err = s.Write(data)
	if err != nil {
		return errors.WithMessage(err, "writing register data")
	}

	return nil
}

// QueryOnChainAddress queries the relay-server for the peerID of a peer given its on-chain address.
func (acc *Account) QueryOnChainAddress(onChainAddr wallet.Address) (*Address, error) {
	id, err := peer.Decode(relayID)
	if err != nil {
		err = errors.WithMessage(err, "decoding peer id of relay server")
		return nil, err
	}

	s, err := acc.NewStream(network.WithAllowLimitedConn(context.Background(), queryProtocol[1:]), id, queryProtocol)
	if err != nil {
		return nil, errors.WithMessage(err, "creating new stream")
	}
	defer s.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	_, err = fmt.Fprintf(rw, "%s\n", onChainAddr)
	if err != nil {
		return nil, errors.WithMessage(err, "writing on-chain address")
	}
	rw.Flush()

	str, _ := rw.ReadString('\n')
	if str == "" {
		return nil, errors.New("empty response from relay server")
	}
	peerIDstr := str[:len(str)-1]
	peerID, err := peer.Decode(peerIDstr)
	if err != nil {
		return nil, errors.WithMessage(err, "decoding peer id")
	}

	return &Address{peerID}, nil
}

func getRelayServerInfo() (*peer.AddrInfo, string, error) {
	id, err := peer.Decode(relayID)
	if err != nil {
		err = errors.WithMessage(err, "decoding peer id of relay server")
		return nil, "", err
	}

	// Get the IP address of the relay server.
	ip, err := net.LookupIP("relay.perun.network")
	if err != nil {
		err = errors.WithMessage(err, "looking up IP address of relay.perun.network")
		return nil, "", err
	}
	relayAddr := "/ip4/" + ip[0].String() + "/tcp/5574"

	relayMultiaddr, err := ma.NewMultiaddr(relayAddr)
	if err != nil {
		err = errors.WithMessage(err, "parsing relay multiadress")
		return nil, "", err
	}

	relayInfo := &peer.AddrInfo{
		ID:    id,
		Addrs: []ma.Multiaddr{relayMultiaddr},
	}

	return relayInfo, relayAddr, nil
}

// keepReservationAlive keeps the reservation alive by periodically renewing it.
func (acc *Account) keepReservationAlive(ctx context.Context, ai peer.AddrInfo) {
	const (
		reserveInterval = 50 * time.Second // slightly less than ReserveTimeout
		initialBackoff  = 2 * time.Second
		maxBackoff      = 1 * time.Minute
	)

	ticker := time.NewTicker(reserveInterval)
	defer ticker.Stop()

	backoff := initialBackoff

	for {
		select {
		case <-ctx.Done():
			err := acc.Close()
			if err != nil {
				panic(err)
			}
			return
		case <-ticker.C:
			newReservation, err := libp2pclient.Reserve(ctx, acc.Host, ai)
			if err != nil {
				log.Printf("keepReservationAlive: reservation failed: %v; retrying in %s", err, backoff)

				timer := time.NewTimer(backoff)
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
					if backoff < maxBackoff {
						backoff *= 2
						if backoff > maxBackoff {
							backoff = maxBackoff
						}
					}
				}
				continue
			}

			acc.reservation = newReservation
			backoff = initialBackoff
			log.Println("keepReservationAlive: reservation successfully renewed")
		}
	}
}
