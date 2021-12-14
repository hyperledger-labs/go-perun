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
	"testing"

	"github.com/stretchr/testify/require"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wallet/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestRequireEqualSigsTX(t *testing.T) {
	prng := pkgtest.Prng(t)
	acc := test.NewRandomAccount(prng)
	data1 := make([]byte, prng.Int63n(50))
	prng.Read(data1)
	data2 := make([]byte, prng.Int63n(50))
	prng.Read(data2)

	sig1, err := acc.SignData(data1)
	require.NoError(t, err)
	sig2, err := acc.SignData(data2)
	require.NoError(t, err)

	equalSigsTableNegative := []struct {
		s1 []wallet.Sig
		s2 []wallet.Sig
	}{
		{[]wallet.Sig{sig1, sig2}, make([]wallet.Sig, 2)},
		{make([]wallet.Sig, 2), []wallet.Sig{sig1, sig2}},
		{make([]wallet.Sig, 5), make([]wallet.Sig, 10)},
		{make([]wallet.Sig, 10), make([]wallet.Sig, 5)},
	}
	equalSigsTablePositive := []struct {
		s1 []wallet.Sig
		s2 []wallet.Sig
	}{
		{nil, nil},
		{nil, make([]wallet.Sig, 10)},
		{make([]wallet.Sig, 10), nil},
		{make([]wallet.Sig, 10), make([]wallet.Sig, 10)},
		{[]wallet.Sig{sig1, sig2}, []wallet.Sig{sig1, sig2}},
		{[]wallet.Sig{sig2}, []wallet.Sig{sig2}},
	}

	tt := pkgtest.NewTester(t)
	for _, _c := range equalSigsTableNegative {
		c := _c
		tt.AssertFatal(func(t pkgtest.T) { requireEqualSigs(t, c.s1, c.s2) })
	}
	for _, c := range equalSigsTablePositive {
		requireEqualSigs(t, c.s1, c.s2)
	}
}
