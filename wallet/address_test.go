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

package wallet_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	_ "perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	peruniotest "perun.network/go-perun/wire/perunio/test"
	pkgtest "polycry.pt/poly-go/test"
)

// TestBackendID is the identifier for the simulated Backend.
const TestBackendID = wallet.BackendID(0)

type testAddresses struct {
	addrs wallet.AddressMapArray
}

func (t *testAddresses) Encode(w io.Writer) error {
	return t.addrs.Encode(w)
}

func (t *testAddresses) Decode(r io.Reader) error {
	return t.addrs.Decode(r)
}

type testAddress struct {
	addrs wallet.AddressDecMap
}

func (t *testAddress) Encode(w io.Writer) error {
	return t.addrs.Encode(w)
}

func (t *testAddress) Decode(r io.Reader) error {
	return t.addrs.Decode(r)
}

func TestAddresses_Serializer(t *testing.T) {
	rng := pkgtest.Prng(t)
	addr := wallettest.NewRandomAddressesMap(rng, 1, TestBackendID)[0]
	peruniotest.GenericSerializerTest(t, &testAddress{addrs: addr})

	addrs := wallettest.NewRandomAddressesMap(rng, 0, TestBackendID)
	peruniotest.GenericSerializerTest(t, &testAddresses{addrs: wallet.AddressMapArray{Addr: addrs}})

	addrs = wallettest.NewRandomAddressesMap(rng, 1, TestBackendID)
	peruniotest.GenericSerializerTest(t, &testAddresses{addrs: wallet.AddressMapArray{Addr: addrs}})

	addrs = wallettest.NewRandomAddressesMap(rng, 5, TestBackendID)
	peruniotest.GenericSerializerTest(t, &testAddresses{addrs: wallet.AddressMapArray{Addr: addrs}})
}

func TestAddrKey_Equal(t *testing.T) {
	rng := pkgtest.Prng(t)
	addrs := wallettest.NewRandomAddressArray(rng, 10, TestBackendID)

	// Test all properties of an equivalence relation.
	for i, a := range addrs {
		for j, b := range addrs {
			// Symmetry.
			require.Equal(t, wallet.Key(a).Equal(b), wallet.Key(b).Equal(a))
			// Test that Equal is equivalent to ==.
			require.Equal(t, wallet.Key(a).Equal(b), wallet.Key(a) == wallet.Key(b))
			// Test that it is not trivially set to true or false.
			require.Equal(t, i == j, wallet.Key(a).Equal(b))
			// Transitivity.
			for _, c := range addrs {
				if wallet.Key(a).Equal(b) && wallet.Key(b).Equal(c) {
					require.True(t, wallet.Key(a).Equal(c))
				}
			}
		}
		// Reflexivity.
		require.True(t, wallet.Key(a).Equal(a))
	}
}

func TestAddrKey(t *testing.T) {
	rng := pkgtest.Prng(t)
	addrs := wallettest.NewRandomAddressArray(rng, 10, TestBackendID)

	for _, a := range addrs {
		// Test that Key and FromKey are dual to each other.
		require.Equal(t, wallet.Key(a), wallet.Key(wallet.FromKey(wallet.Key(a))))
		// Test that FromKey returns the correct Address.
		require.True(t, a.Equal(wallet.FromKey(wallet.Key(a))))
	}
}

func TestCloneAddress(t *testing.T) {
	rng := pkgtest.Prng(t)
	addr := wallettest.NewRandomAddress(rng, TestBackendID)
	addr0 := wallet.CloneAddress(addr)
	require.Equal(t, addr, addr0)
	require.NotSame(t, addr, addr0)
}

func TestCloneAddresses(t *testing.T) {
	rng := pkgtest.Prng(t)
	addrs := wallettest.NewRandomAddressArray(rng, 3, TestBackendID)
	addrs0 := wallet.CloneAddresses(addrs)
	require.Equal(t, addrs, addrs0)
	require.NotSame(t, &addrs, &addrs0)

	for i, a := range addrs {
		require.NotSame(t, a, addrs0[i])
	}
}
