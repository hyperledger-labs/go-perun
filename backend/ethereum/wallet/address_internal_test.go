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

package wallet

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	pkgtest "polycry.pt/poly-go/test"
)

func TestAsWalletAddr(t *testing.T) {
	t.Run("non-zero-value", func(t *testing.T) {
		rng := pkgtest.Prng(t)
		var commonAddr common.Address
		rng.Read(commonAddr[:])

		ethAddr := AsWalletAddr(commonAddr)
		require.Equal(t, commonAddr.String(), ethAddr.String())
	})
	t.Run("zero-value", func(t *testing.T) {
		var commonAddr common.Address

		ethAddr := AsWalletAddr(commonAddr)
		require.Equal(t, commonAddr.String(), ethAddr.String())
	})
}

func TestAddressOrdering(t *testing.T) {
	const LT, EQ, GT = -1, 0, 1
	cases := []struct {
		addr     [2]Address
		expected int
	}{
		{[2]Address{{1, 1}, {2, 1}}, LT},
		{[2]Address{{1, 1}, {1, 2}}, LT},
		{[2]Address{{0, 1}, {1, 2}}, LT},
		{[2]Address{{2, 1}, {1, 1}}, GT},
		{[2]Address{{2, 2}, {2, 1}}, GT},
		{[2]Address{{2, 1}, {1, 0}}, GT},
		{[2]Address{{1, 1}, {1, 1}}, EQ},
	}

	for _, c := range cases {
		require.Equal(t, c.expected, c.addr[0].Cmp(&c.addr[1]))
	}
}
