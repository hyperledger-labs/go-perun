// Copyright 2020 - See NOTICE file for copyright holders.
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
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	pkgtest "polycry.pt/poly-go/test"

	"perun.network/go-perun/pkg/io/test"
)

func TestGenericMarshaler(t *testing.T) {
	rng := pkgtest.Prng(t)
	for n := 0; n < 10; n++ {
		test.GenericMarshalerTest(t, NewRandomAddress(rng))
	}
}

func TestAddressMarshalling(t *testing.T) {
	dest := [AddressBinaryLen]byte{}
	t.Run("full length", func(t *testing.T) {
		for i := range dest {
			dest[i] = byte(i)
		}

		addr := &Address{
			X: new(big.Int).SetBytes(dest[:ElemLen]),
			Y: new(big.Int).SetBytes(dest[ElemLen:]),
		}
		result := addr.ByteArray()
		assert.Equal(t, result[:], dest[:])
	})

	t.Run("half length", func(t *testing.T) {
		const zeros = 5
		for i := 0; i < zeros; i++ {
			dest[i] = 0
			dest[i+ElemLen] = 0
		}

		addr := &Address{
			X: new(big.Int).SetBytes(dest[:ElemLen]),
			Y: new(big.Int).SetBytes(dest[ElemLen:]),
		}
		result := addr.ByteArray()
		assert.Equal(t, result[:], dest[:])
	})
}

func TestAddressOrdering(t *testing.T) {
	const EQ, LT, GT = 0, -1, 1
	rng := pkgtest.Prng(t)

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
