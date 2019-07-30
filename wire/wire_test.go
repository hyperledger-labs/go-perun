// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"io"
	"math/big"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/io/test"
	peruntest "perun.network/go-perun/pkg/test"
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

func TestByteSlice(t *testing.T) {
	var v1, v2, v3 ByteSlice = []byte{}, []byte{255}, []byte{1, 2, 3, 4, 5, 6}
	testByteSlices(t, v1, v2, v3)
}

func TestBigInt(t *testing.T) {
	var v1 = BigInt{big.NewInt(123456)}
	var v2 = BigInt{big.NewInt(1)}
	var v3 = BigInt{big.NewInt(0)}
	test.GenericSerializableTest(t, &v1, &v2, &v3)
}

func TestInvalidBigInt(t *testing.T) {
	a := assert.New(t)
	// Test integers that are too big
	bytes := make([]byte, maxBigIntLength+1)
	bytes[0] = 1
	_big := big.NewInt(1).SetBytes(bytes)
	var tooBig = BigInt{_big}
	r, w := io.Pipe()

	a.NotNil(tooBig.Encode(w), "encoding of a big integer that is too big should fail")

	go func() {
		w.Write([]byte{uint8(len(bytes))})
	}()

	var result BigInt
	a.NotNil(result.Decode(r), "decoding of an integer that is too big should fail")

	// Test not sending value, only length
	go func() {
		w.Write([]byte{10})
		w.Close()
	}()

	a.NotNil(result.Decode(r), "decoding after sender only send length should fail")
}

func TestWrongTypes(t *testing.T) {
	r, w := io.Pipe()

	values := []interface{}{
		errors.New(""),
		int8(1),
		byte(7),
		float32(1.2),
		float64(1.3),
		complex(1, 2),
		complex128(1),
	}

	peruntest.CheckPanic(func() { Encode(w, values...) })

	d := make([]interface{}, len(values))
	for i, v := range values {
		d[i] = reflect.New(reflect.TypeOf(v)).Interface()
	}

	peruntest.CheckPanic(func() { Decode(r, d...) })
	// Assert that SizeType can
	if unsafe.Sizeof(maxBigIntLength) != unsafe.Sizeof(uint8(0)) {
		t.Error("maxBigIntLength must have type uint8")
	}
}

func TestEncodeDecode(t *testing.T) {
	a := assert.New(t)
	r, w := io.Pipe()

	values := []interface{}{
		true,
		uint16(0x1234),
		uint32(0x123567),
		uint64(0x1234567890123456),
		int16(0x1234),
		int32(0x123567),
		int64(0x1234567890123456),
		// The time has to be constructed this way, because otherwise DeepEqual fails.
		time.Unix(0, time.Now().UnixNano()),
	}

	go func() {
		a.Nil(Encode(w, values...), "failed to encode values")
	}()

	d := make([]interface{}, len(values))
	for i, v := range values {
		d[i] = reflect.New(reflect.TypeOf(v)).Interface()
	}

	a.Nil(Decode(r, d...), "failed to decode values")

	for i, v := range values {
		if !reflect.DeepEqual(reflect.ValueOf(d[i]).Elem().Interface(), v) {
			t.Errorf("%dth values are not the same: %T %v, %T %v", i, v, v, d[i], d[i])
		}
	}
}

func testByteSlices(t *testing.T, serial ...ByteSlice) {
	a := assert.New(t)
	for i, v := range serial {
		r, w := io.Pipe()

		d := make([]byte, len(v))
		dest := ByteSlice(d)
		go func(v ByteSlice) {
			a.Nil(v.Encode(w), "failed to encode element")
		}(v)

		a.Nil(dest.Decode(r), "failed to decode element")

		if !reflect.DeepEqual(v, dest) {
			t.Errorf("encoding and decoding the %dth element (%T) resulted in different value: %v, %v", i, v, reflect.ValueOf(v).Elem(), dest)
		}
	}

	for _, v := range serial {
		r, w := io.Pipe()
		w.Close()
		a.NotNil(v.Encode(w), "encoding on closed writer should fail, but does not.")

		r.Close()
		a.False(v.Decode(r) == nil && len(v) != 0, "decoding on closed reader should fail, but does not.")
	}
}
