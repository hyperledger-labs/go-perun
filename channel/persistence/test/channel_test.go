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

package test

import (
	"math/rand"
	"testing"

	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wallet"
)

func TestRequireEqualSigsTX(t *testing.T) {
	prng := test.Prng(t)
	var equalSigsTableNegative = []struct {
		s1 []wallet.Sig
		s2 []wallet.Sig
	}{
		{initSigSlice(10, prng), make([]wallet.Sig, 10)},
		{make([]wallet.Sig, 10), initSigSlice(10, prng)},
		{[]wallet.Sig{{4, 3, 2, 1}}, []wallet.Sig{{1, 2, 3, 4}}},
		{make([]wallet.Sig, 5), make([]wallet.Sig, 10)},
		{make([]wallet.Sig, 10), make([]wallet.Sig, 5)},
	}
	var equalSigsTablePositive = []struct {
		s1 []wallet.Sig
		s2 []wallet.Sig
	}{
		{nil, nil},
		{nil, make([]wallet.Sig, 10)},
		{make([]wallet.Sig, 10), nil},
		{make([]wallet.Sig, 10), make([]wallet.Sig, 10)},
		{[]wallet.Sig{{1, 2, 3, 4}}, []wallet.Sig{{1, 2, 3, 4}}},
	}

	tt := test.NewTester(t)
	for _, _c := range equalSigsTableNegative {
		c := _c
		tt.AssertFatal(func(t test.T) { requireEqualSigs(t, c.s1, c.s2) })
	}
	for _, c := range equalSigsTablePositive {
		requireEqualSigs(t, c.s1, c.s2)
	}
}

func initSigSlice(length int, prng *rand.Rand) []wallet.Sig {
	s := make([]wallet.Sig, length)
	for i := range s {
		s[i] = make(wallet.Sig, 32)
		for j := 0; j < 32; j++ {
			s[i][j] = byte(prng.Int())
		}
	}
	return s
}
