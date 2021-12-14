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

package test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/wire/perunio"
)

// TestAddress runs a test suite designed to test the general functionality of
// an address implementation.
func TestAddress(t *testing.T, s *Setup) { //nolint:revive // `test.Test...` stutters, but we accept that here.
	null := s.ZeroAddress
	addr := s.Backend.NewAddress()
	assert.NoError(t, addr.UnmarshalBinary(s.AddressMarshalled), "Byte deserialization of address should work")

	// Test Address.String.
	nullString := null.String()
	addrString := addr.String()
	assert.Greater(t, len(nullString), 0)
	assert.Greater(t, len(addrString), 0)
	assert.NotEqual(t, addrString, nullString)

	// Test Address.Equals.
	assert.False(t, addr.Equal(null), "Expected inequality of zero, nonzero address")
	assert.True(t, null.Equal(null), "Expected equality of zero address to itself") //nolint:gocritic

	// Test Address.Cmp.
	assert.Positive(t, addr.Cmp(null), "Expected addr > zero")
	assert.Zero(t, null.Cmp(null), "Expected zero = zero") //nolint:gocritic
	assert.Negative(t, null.Cmp(addr), "Expected null < addr")

	// Test Address.Bytes.
	addrBytes, err := addr.MarshalBinary()
	assert.NoError(t, err, "Marshaling address should not error")
	nullBytes, err := null.MarshalBinary()
	assert.NoError(t, err, "Marshaling zero address should not error")
	assert.False(t, bytes.Equal(addrBytes, nullBytes), "Expected inequality of byte representations of nonzero and zero address")

	// a.Equal(Decode(Encode(a)))
	t.Run("Serialize Equal Test", func(t *testing.T) {
		buff := new(bytes.Buffer)
		require.NoError(t, perunio.Encode(buff, addr))
		addr2 := s.Backend.NewAddress()
		err := perunio.Decode(buff, addr2)
		require.NoError(t, err)

		assert.True(t, addr.Equal(addr2))
	})
}
