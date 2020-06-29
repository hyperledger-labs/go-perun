// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"math/rand"
	"testing"

	"perun.network/go-perun/wallet"
)

type mockEqualT struct {
	*testing.T
	Equal bool
}

// Errorf formats an error message
func (m *mockEqualT) Errorf(format string, args ...interface{}) {
	return
}

// FailNow marks the function as having failed
func (m *mockEqualT) FailNow() {
	m.Equal = false
}

func (m *mockEqualT) reset() {
	m.Equal = true
}

func TestRequireEqualSigsTX(t *testing.T) {
	var equalSigsTable = []struct {
		s1    []wallet.Sig
		s2    []wallet.Sig
		Equal bool
	}{
		{nil, nil, true},
		{nil, make([]wallet.Sig, 10), true},
		{make([]wallet.Sig, 10), nil, true},
		{make([]wallet.Sig, 10), make([]wallet.Sig, 10), true},
		{initSigSlice(10), make([]wallet.Sig, 10), false},
		{make([]wallet.Sig, 10), initSigSlice(10), false},
		{[]wallet.Sig{{1, 2, 3, 4}}, []wallet.Sig{{1, 2, 3, 4}}, true},
		{[]wallet.Sig{{4, 3, 2, 1}}, []wallet.Sig{{1, 2, 3, 4}}, false},
		{make([]wallet.Sig, 5), make([]wallet.Sig, 10), false},
		{make([]wallet.Sig, 10), make([]wallet.Sig, 5), false},
	}

	wrapT := &mockEqualT{T: t, Equal: true}
	for i, c := range equalSigsTable {
		requireEqualSigs(wrapT, c.s1, c.s2)
		if got := wrapT.Equal; got != c.Equal {
			t.Errorf("Case: %v => wanted Equal to be %#v, but got: %#v", i, c.Equal, got)
		}
		wrapT.reset()
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
