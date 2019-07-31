// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"testing"

	"perun.network/go-perun/pkg/test"
)

type mockBackend struct {
	test.WrapMock
}

// channel.Backend interface

func (m *mockBackend) ChannelID(p *Params) ID {
	m.AssertWrapped()
	return Zero
}

// compile-time check that mockBackend imlements Backend
var _ Backend = (*mockBackend)(nil)

// TestGlobalBackend tests all global backend wrappers
func TestGlobalBackend(t *testing.T) {
	b := &mockBackend{test.NewWrapMock(t)}
	SetBackend(b)
	ChannelID(nil)
	b.AssertCalled()
}
