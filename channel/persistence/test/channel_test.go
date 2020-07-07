// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"math/rand"
	"testing"

	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wallet"
)

func TestRequireEqualSigsTX(t *testing.T) {
	var equalSigsTableNegative = []struct {
		s1 []wallet.Sig
		s2 []wallet.Sig
	}{
		{initSigSlice(10), make([]wallet.Sig, 10)},
		{make([]wallet.Sig, 10), initSigSlice(10)},
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
	for _, c := range equalSigsTableNegative {
		tt.AssertFatal(func(t test.T) { requireEqualSigs(t, c.s1, c.s2) })
	}
	for _, c := range equalSigsTablePositive {
		requireEqualSigs(t, c.s1, c.s2)
	}
}

func initSigSlice(length int) []wallet.Sig {
	s := make([]wallet.Sig, length)
	for i := range s {
		s[i] = make(wallet.Sig, 32)
		for j := 0; j < 32; j++ {
			s[i][j] = byte(rand.Int())
		}
	}
	return s
}
