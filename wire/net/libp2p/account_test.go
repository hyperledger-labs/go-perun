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

package libp2p_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	wtest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire/net/libp2p"
	pkgtest "polycry.pt/poly-go/test"
)

func TestNewAccount(t *testing.T) {
	rng := pkgtest.Prng(t)
	acc := libp2p.NewRandomAccount(rng)
	assert.NotNil(t, acc)
	defer acc.Close()
}

func TestAddressBookRegister(t *testing.T) {
	rng := pkgtest.Prng(t)
	acc := libp2p.NewRandomAccount(rng)
	assert.NotNil(t, acc)
	defer acc.Close()

	onChainAddr := wtest.NewRandomAddress(rng, channel.TestBackendID)

	err := acc.RegisterOnChainAddress(onChainAddr)
	assert.NoError(t, err)
}

func TestAddressBookRegisterEmptyAddress(t *testing.T) {
	rng := pkgtest.Prng(t)
	acc := libp2p.NewRandomAccount(rng)
	assert.NotNil(t, acc)

	defer acc.Close()

	var nilAddr wallet.Address
	err := acc.RegisterOnChainAddress(nilAddr)
	assert.Error(t, err)
}

func TestAddressBookDeregister(t *testing.T) {
	rng := pkgtest.Prng(t)
	acc := libp2p.NewRandomAccount(rng)
	assert.NotNil(t, acc)
	defer acc.Close()

	onChainAddr := wtest.NewRandomAddress(rng, channel.TestBackendID)

	err := acc.RegisterOnChainAddress(onChainAddr)
	assert.NoError(t, err)

	err = acc.DeregisterOnChainAddress(onChainAddr)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Trying to query it again will fail
	_, err = acc.QueryOnChainAddress(onChainAddr)
	assert.Error(t, err)
}

func TestAddressBookDeregisterPeer(t *testing.T) {
	rng := pkgtest.Prng(t)
	acc := libp2p.NewRandomAccount(rng)
	assert.NotNil(t, acc)
	defer acc.Close()

	peer := libp2p.NewRandomAccount(rng)
	assert.NotNil(t, peer)
	defer peer.Close()

	onChainAddr := wtest.NewRandomAddress(rng, channel.TestBackendID)
	peerOnChainAddr := wtest.NewRandomAddress(rng, channel.TestBackendID)

	err := acc.RegisterOnChainAddress(onChainAddr)
	assert.NoError(t, err)

	time.Sleep(1 * time.Millisecond)

	err = peer.RegisterOnChainAddress(peerOnChainAddr)
	assert.NoError(t, err)

	err = acc.DeregisterOnChainAddress(onChainAddr)
	assert.NoError(t, err)

	// Trying to deregister the peer's address will not fail, but the server will not allow it.
	err = acc.DeregisterOnChainAddress(peerOnChainAddr)
	assert.NoError(t, err)

	// Trying to query it again will be okay
	peerID, err := acc.QueryOnChainAddress(peerOnChainAddr)
	assert.NoError(t, err)

	addr := peer.Address()
	assert.Equal(t, peerID, addr)

	err = peer.DeregisterOnChainAddress(peerOnChainAddr)
	assert.NoError(t, err)
}

func TestAddressBookQuery_Fail(t *testing.T) {
	rng := pkgtest.Prng(t)
	acc := libp2p.NewRandomAccount(rng)
	assert.NotNil(t, acc)
	defer acc.Close()

	onChainAddr := wtest.NewRandomAddress(rng, channel.TestBackendID)

	_, err := acc.QueryOnChainAddress(onChainAddr)
	assert.Error(t, err)
}

func TestAddressBookQuery(t *testing.T) {
	rng := pkgtest.Prng(t)
	acc := libp2p.NewRandomAccount(rng)
	assert.NotNil(t, acc)
	defer acc.Close()

	onChainAddr := wtest.NewRandomAddress(rng, channel.TestBackendID)

	err := acc.RegisterOnChainAddress(onChainAddr)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)
	peerID, err := acc.QueryOnChainAddress(onChainAddr)
	assert.NoError(t, err)

	addr := acc.Address()
	assert.Equal(t, peerID, addr)

	err = acc.DeregisterOnChainAddress(onChainAddr)
	assert.NoError(t, err)
}

func TestAddressBookQueryPeer(t *testing.T) {
	rng := pkgtest.Prng(t)
	acc := libp2p.NewRandomAccount(rng)
	assert.NotNil(t, acc)
	defer acc.Close()

	peer := libp2p.NewRandomAccount(rng)
	assert.NotNil(t, peer)
	defer peer.Close()

	onChainAddr := wtest.NewRandomAddress(rng, channel.TestBackendID)
	peerOnChainAddr := wtest.NewRandomAddress(rng, channel.TestBackendID)

	err := acc.RegisterOnChainAddress(onChainAddr)
	assert.NoError(t, err)

	err = peer.RegisterOnChainAddress(peerOnChainAddr)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	peerID, err := acc.QueryOnChainAddress(peerOnChainAddr)
	assert.NoError(t, err)

	addr := peer.Address()
	assert.Equal(t, peerID, addr)

	err = acc.DeregisterOnChainAddress(onChainAddr)
	assert.NoError(t, err)

	err = acc.DeregisterOnChainAddress(peerOnChainAddr)
	assert.NoError(t, err)
}

func TestAddressBookRegisterQueryMultiple(t *testing.T) {
	rng := pkgtest.Prng(t)
	acc := libp2p.NewRandomAccount(rng)
	assert.NotNil(t, acc)
	defer acc.Close()

	onChainAddr := wtest.NewRandomAddress(rng, channel.TestBackendID)
	onChainAddr2 := wtest.NewRandomAddress(rng, channel.TestBackendID)

	err := acc.RegisterOnChainAddress(onChainAddr)
	assert.NoError(t, err)

	err = acc.RegisterOnChainAddress(onChainAddr2)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	accID, err := acc.QueryOnChainAddress(onChainAddr)
	assert.NoError(t, err)

	accID2, err := acc.QueryOnChainAddress(onChainAddr2)
	assert.NoError(t, err)

	addr := acc.Address()
	assert.Equal(t, accID, addr)
	assert.Equal(t, accID2, addr)

	// Clean up
	err = acc.DeregisterOnChainAddress(onChainAddr)
	assert.NoError(t, err)

	err = acc.DeregisterOnChainAddress(onChainAddr2)
	assert.NoError(t, err)
}

func TestNewAccountFromPrivateKey(t *testing.T) {
	rng := pkgtest.Prng(t)
	acc := libp2p.NewRandomAccount(rng)
	assert.NotNil(t, acc)

	defer acc.Close()

	keyBytes, err := acc.MarshalPrivateKey()
	assert.NoError(t, err)

	acc2, err := libp2p.NewAccountFromPrivateKeyBytes(keyBytes)
	assert.NoError(t, err)

	defer acc2.Close()

	assert.NotNil(t, acc2)
	assert.Equal(t, acc.ID(), acc2.ID())
	assert.Equal(t, acc.Address(), acc2.Address())
}
