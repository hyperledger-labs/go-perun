// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

// +build wrap_test

package channel

import (
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wallet"
)

type mockBackend struct {
	test.WrapMock
}

// channel.Backend interface

func (m *mockBackend) ChannelID(*Params) ID {
	m.AssertWrapped()
	return Zero
}

func (m *mockBackend) Sign(wallet.Account, *Params, *State) (wallet.Sig, error) {
	m.AssertWrapped()
	return nil, nil
}

func (m *mockBackend) Verify(wallet.Address, *Params, *State, wallet.Sig) (bool, error) {
	m.AssertWrapped()
	return false, nil
}

func (m *mockBackend) DecodeAsset(io.Reader) (Asset, error) {
	m.AssertWrapped()
	return nil, nil
}

// compile-time check that mockBackend implements Backend
var _ Backend = (*mockBackend)(nil)

// TestGlobalBackend tests all global backend wrappers
func TestGlobalBackend(t *testing.T) {
	b := &mockBackend{test.NewWrapMock(t)}
	SetBackend(b)

	ChannelID(nil)
	b.AssertCalled()

	Sign(nil, nil, nil)
	b.AssertCalled()

	Verify(nil, nil, nil, nil)
	b.AssertCalled()

	DecodeAsset(nil)
	b.AssertCalled()
}

func TestMaxConstants(t *testing.T) {
	assert.LessOrEqual(t, MaxNumAssets, math.MaxUint16, "MaxNumAssets must not be greater than math.MaxUint16")
	assert.LessOrEqual(t, MaxNumParts, math.MaxUint16, "MaxNumParts must not be greater than math.MaxUint16")
	assert.LessOrEqual(t, MaxNumSubAllocations, math.MaxUint16, "MaxNumSubAllocations must not be greater than math.MaxUint16")
}
