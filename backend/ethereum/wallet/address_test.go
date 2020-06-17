// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wallet

import (
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestAsWalletAddr(t *testing.T) {
	t.Run("non-zero-value", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1929))
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
