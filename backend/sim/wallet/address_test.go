// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wallet

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/pkg/test"
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

func TestAddressOrdering(t *testing.T) {
	const EQ, LT, GT = 0, -1, 1
	rng := test.Prng(t)

	type addrTest struct {
		addr     [2]*Address
		expected int
	}
	var cases []addrTest
	for i := 0; i < 10; i++ {
		addr := NewRandomAddress(rng)
		ltX := CloneAddr(addr)
		ltY := CloneAddr(addr)
		gtX := CloneAddr(addr)
		gtY := CloneAddr(addr)
		ltXgtY := CloneAddr(addr)
		gtXltY := CloneAddr(addr)
		ltX.X.Sub(ltX.X, big.NewInt(0x1))
		ltY.Y.Sub(ltY.Y, big.NewInt(0x1))
		gtX.X.Add(gtX.X, big.NewInt(0x1))
		gtY.Y.Add(gtX.Y, big.NewInt(0x1))
		ltXgtY.X.Sub(ltXgtY.X, big.NewInt(0x1))
		ltXgtY.Y.Add(ltXgtY.Y, big.NewInt(0x1))
		gtXltY.X.Add(gtXltY.X, big.NewInt(0x1))
		gtXltY.Y.Sub(gtXltY.Y, big.NewInt(0x1))
		addrlt1 := addrTest{
			addr:     [2]*Address{ltX, addr},
			expected: LT,
		}
		addrlt2 := addrTest{
			addr:     [2]*Address{ltY, addr},
			expected: LT,
		}
		addrgt1 := addrTest{
			addr:     [2]*Address{gtX, addr},
			expected: GT,
		}
		addrgt2 := addrTest{
			addr:     [2]*Address{gtY, addr},
			expected: GT,
		}
		addreq := addrTest{
			addr:     [2]*Address{addr, addr},
			expected: EQ,
		}
		addrltXgtY := addrTest{
			addr:     [2]*Address{ltXgtY, addr},
			expected: LT,
		}
		addrgtXltY := addrTest{
			addr:     [2]*Address{gtXltY, addr},
			expected: GT,
		}
		cases = append(cases, addrltXgtY, addrgtXltY, addrlt1, addrlt2, addrgt1, addrgt2, addreq)
	}

	for _, c := range cases {
		assert.Equal(t, c.expected, c.addr[0].Cmp(c.addr[1]),
			"comparison of addresses had unexpected result")
	}
}

func CloneAddr(addr *Address) *Address {
	cloneX, cloneY := big.NewInt(0).Set(addr.X), big.NewInt(0).Set(addr.Y)
	return &Address{
		Curve: addr.Curve,
		X:     cloneX,
		Y:     cloneY,
	}
}
