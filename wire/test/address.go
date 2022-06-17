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

package test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/perunio"
	pkgtest "polycry.pt/poly-go/test"
)

// TestAddressImplementation runs a test suite designed to test the general
// functionality of an address implementation.
//nolint:revive // The function name `test.Test...` stutters, but it is OK in this special case.
func TestAddressImplementation(t *testing.T, newAddress wire.NewAddressFunc, newRandomAddress NewRandomAddressFunc) {
	rng := pkgtest.Prng(t)
	require, assert := require.New(t), assert.New(t)
	addr := newRandomAddress(rng)

	// Test Address.MarshalBinary and UnmarshalBinary.
	data, err := addr.MarshalBinary()
	assert.NoError(err)
	assert.NoError(addr.UnmarshalBinary(data), "Byte deserialization of address should work")

	// Test Address.Equals.
	null := newAddress()
	assert.False(addr.Equal(null), "Expected inequality of zero, nonzero address")
	assert.True(null.Equal(null), "Expected equality of zero address to itself") //nolint:gocritic

	// Test Address.Cmp.
	assert.Positive(addr.Cmp(null), "Expected addr > zero")
	assert.Zero(null.Cmp(null), "Expected zero = zero") //nolint:gocritic
	assert.Negative(null.Cmp(addr), "Expected null < addr")

	// Test Address.Bytes.
	addrBytes, err := addr.MarshalBinary()
	assert.NoError(err, "Marshaling address should not error")
	nullBytes, err := null.MarshalBinary()
	assert.NoError(err, "Marshaling zero address should not error")
	assert.False(bytes.Equal(addrBytes, nullBytes), "Expected inequality of byte representations of nonzero and zero address")

	// a.Equal(Decode(Encode(a)))
	t.Run("Serialize Equal Test", func(t *testing.T) {
		buff := new(bytes.Buffer)
		require.NoError(perunio.Encode(buff, addr))
		addr2 := newAddress()
		err := perunio.Decode(buff, addr2)
		require.NoError(err)

		assert.True(addr.Equal(addr2))
	})
}
