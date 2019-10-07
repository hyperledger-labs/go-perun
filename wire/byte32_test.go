// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"testing"

	"perun.network/go-perun/pkg/io/test"
)

func TestByte32(t *testing.T) {
	v1, v2 := Byte32{}, Byte32{}
	for i := byte(0); i < 32; i++ {
		v2[i] = i + 1 // all non-zero [32]byte
	}
	test.GenericSerializableTest(t, &v1, &v2)
}
