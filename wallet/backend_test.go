// Copyright 2019 - See NOTICE file for copyright holders.
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

// +build wrap_test

package wallet

import (
	"io"
	"testing"

	"perun.network/go-perun/pkg/test"
)

type mockBackend struct {
	test.WrapMock
}

// wallet.Backend interface

// DecodeAddress reads and decodes an address from an io.Writer
func (m *mockBackend) DecodeAddress(io.Reader) (Address, error) {
	m.AssertWrapped()
	return nil, nil
}

func (m *mockBackend) DecodeSig(io.Reader) (Sig, error) {
	m.AssertWrapped()
	return nil, nil
}

func (m *mockBackend) VerifySignature([]byte, Sig, Address) (bool, error) {
	m.AssertWrapped()
	return false, nil
}

// compile-time check that mockBackend imlements Backend
var _ Backend = (*mockBackend)(nil)

// TestGlobalBackend tests all global backend wrappers
func TestGlobalBackend(t *testing.T) {
	b := &mockBackend{test.NewWrapMock(t)}
	SetBackend(b)
	DecodeAddress(nil)
	b.AssertCalled()
	DecodeSig(nil)
	b.AssertCalled()
	VerifySignature(nil, nil, nil)
	b.AssertCalled()
}
