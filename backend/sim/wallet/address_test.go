// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wallet

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddress_ByteArray(t *testing.T) {
	dest := [64]byte{}
	t.Run("full length", func(t *testing.T) {
		for i := range dest {
			dest[i] = byte(i)
		}

		addr := &Address{
			X: new(big.Int).SetBytes(dest[:32]),
			Y: new(big.Int).SetBytes(dest[32:])}
		result := addr.ByteArray()
		assert.Equal(t, result[:], dest[:])
	})

	t.Run("half length", func(t *testing.T) {
		const zeros = 5
		for i := 0; i < zeros; i++ {
			dest[i] = 0
			dest[i+32] = 0
		}

		addr := &Address{
			X: new(big.Int).SetBytes(dest[:32]),
			Y: new(big.Int).SetBytes(dest[32:])}
		result := addr.ByteArray()
		assert.Equal(t, result[:], dest[:])
	})
}
