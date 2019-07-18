// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"testing"

	"perun.network/go-perun/pkg/io/test"
)

func TestBool(t *testing.T) {
	var tr Bool = true
	var fa Bool = false
	test.GenericSerializableTest(t, &tr, &fa)
}

func TestInt16(t *testing.T) {
	var v1, v2, v3 Int16 = 0, -0x1117, 0x4334
	test.GenericSerializableTest(t, &v1, &v2, &v3)
}

func TestInt32(t *testing.T) {
	var v1, v2, v3 Int32 = 0, -0x11223344, 0x34251607
	test.GenericSerializableTest(t, &v1, &v2, &v3)
}

func TestInt64(t *testing.T) {
	var v1, v2, v3 Int64 = 0, -0x1234567890123456, 0x5920838589479478
	test.GenericSerializableTest(t, &v1, &v2, &v3)
}

func TestTime(t *testing.T) {
	var v1, v2, v3 Time = 0, 0x3478534567898762, 0x7975089975789098
	test.GenericSerializableTest(t, &v1, &v2, &v3)
}
