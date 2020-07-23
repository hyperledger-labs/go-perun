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

package io_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/io"
)

// TestEqualEncoding tests EqualEncoding
func TestEqualEncoding(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	a := make(io.ByteSlice, 10)
	b := make(io.ByteSlice, 10)
	c := make(io.ByteSlice, 12)

	rng.Read(a)
	rng.Read(b)
	rng.Read(c)
	c2 := io.ByteSlice(c)

	tests := []struct {
		a         io.Encoder
		b         io.Encoder
		shouldOk  bool
		shouldErr bool
		name      string
	}{
		{a, nil, false, true, "one Encoder set to nil"},
		{nil, a, false, true, "one Encoder set to nil"},
		{io.Encoder(nil), b, false, true, "one Encoder set to nil"},
		{b, io.Encoder(nil), false, true, "one Encoder set to nil"},

		{nil, nil, true, false, "both Encoders set to nil"},
		{io.Encoder(nil), io.Encoder(nil), true, false, "both Encoders set to nil"},

		{a, a, true, false, "same Encoders"},
		{a, &a, true, false, "same Encoders"},
		{&a, a, true, false, "same Encoders"},
		{&a, &a, true, false, "same Encoders"},

		{c, c2, true, false, "different Encoders and same content"},

		{a, b, false, false, "different Encoders and different content"},
		{a, c, false, false, "different Encoders and different content"},
	}

	for _, tt := range tests {
		ok, err := io.EqualEncoding(tt.a, tt.b)

		assert.Equalf(t, ok, tt.shouldOk, "EqualEncoding with %s should return %t as bool but got: %t", tt.name, tt.shouldOk, ok)
		assert.Falsef(t, (err == nil) && tt.shouldErr, "EqualEncoding with %s should return an error but got nil", tt.name)
		assert.Falsef(t, (err != nil) && !tt.shouldErr, "EqualEncoding with %s should return nil as error but got: %s", tt.name, err)
	}
}
