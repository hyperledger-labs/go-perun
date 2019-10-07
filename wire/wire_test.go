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

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	peruntest "perun.network/go-perun/pkg/test"
)

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

	d := make([]interface{}, len(values))
	for i, v := range values {
		panics, _ := peruntest.CheckPanic(func() { Encode(w, v) })
		assert.True(t, panics, "Encode() must panic on invalid type %T", v)

		d[i] = reflect.New(reflect.TypeOf(v)).Interface()
		panics, _ = peruntest.CheckPanic(func() { Decode(r, d[i]) })
		assert.True(t, panics, "Decode() must panic on invalid type %T", v)
	}

	peruntest.CheckPanic(func() { Decode(r, d...) })
}

func TestEncodeDecode(t *testing.T) {
	a := assert.New(t)
	r, w := io.Pipe()

	longInt, _ := new(big.Int).SetString("12345671823897123798901234561234567890", 16)
	var byte32 [32]byte
	for i := byte(0); i < 32; i++ {
		byte32[i] = i + 1
	}
	byteSlice := []byte{0, 1, 2, 4, 8, 0x10, 0x20, 0x40, 0x80}
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
		big.NewInt(0x1234567890123456),
		longInt,
		byte32,
		byteSlice,
	}

	go func() {
		a.Nil(Encode(w, values...), "failed to encode values")
	}()

	d := make([]interface{}, len(values))
	for i, v := range values {
		if b, ok := v.([]byte); ok {
			// destination byte slice has to be of correct size
			_d := make([]byte, len(b))
			d[i] = &_d
		} else {
			d[i] = reflect.New(reflect.TypeOf(v)).Interface()
		}
	}

	a.Nil(Decode(r, d...), "failed to decode values")

	for i, v := range values {
		if !reflect.DeepEqual(reflect.ValueOf(d[i]).Elem().Interface(), v) {
			t.Errorf("%dth values are not the same: %T %v, %T %v", i, v, v, d[i], d[i])
		}
	}
}
