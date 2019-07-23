// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"io"
	"reflect"
	"testing"
	"time"

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
	v1, v2, v3 := Time{0}, Time{0x3478534567898762}, Time{0x7975089975789098}
	test.GenericSerializableTest(t, &v1, &v2, &v3)
}

func TestEncodeDecode(t *testing.T) {
	r, w := io.Pipe()

	values := []interface{}{
		true,
		uint16(0x1234),
		uint32(0x123567),
		uint64(0x1234567890123456),
		// The time has to be constructed this way, because otherwise DeepEqual fails.
		time.Unix(0, time.Now().UnixNano()),
	}

	go func() {
		if err := Encode(w, values...); err != nil {
			t.Errorf("failed to write values: %+v", err)
		}
	}()

	d := make([]interface{}, len(values))
	for i, v := range values {
		d[i] = reflect.New(reflect.TypeOf(v)).Interface()
	}
	if err := Decode(r, d...); err != nil {
		t.Errorf("failed to read values: %+v", err)
	}

	for i, v := range values {
		if !reflect.DeepEqual(reflect.ValueOf(d[i]).Elem().Interface(), v) {
			t.Errorf("%dth values are not the same: %T %v, %T %v", i, v, v, d[i], d[i])
		}
	}
}
