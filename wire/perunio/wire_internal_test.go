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

package perunio

import (
	"io"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	polytest "polycry.pt/poly-go/test"
)

func TestWrongTypes(t *testing.T) {
	r, w := io.Pipe()

	values := []interface{}{
		errors.New(""),
		float32(1.2),
		float64(1.3),
		complex(1, 2),
		complex128(1),
	}

	d := make([]interface{}, len(values))
	for _i, _v := range values {
		v := _v
		i := _i
		panics, _ := polytest.CheckPanic(func() { _ = Encode(w, v) })
		assert.True(t, panics, "Encode() must panic on invalid type %T", v)

		d[i] = reflect.New(reflect.TypeOf(v)).Interface()
		panics, _ = polytest.CheckPanic(func() { _ = Decode(r, d[i]) })
		assert.True(t, panics, "Decode() must panic on invalid type %T", v)
	}

	polytest.CheckPanic(func() { _ = Decode(r, d...) })
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
		byte(0xB0),
		int8(-127),
		uint8(0xB0),
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
		ByteSlice{5, 6, 8, 3, 4, 5, 6},
		"perun",
	}

	go func() {
		a.NoError(Encode(w, values...), "failed to encode values")
	}()

	d := make([]interface{}, len(values))
	for i, v := range values {
		if b, ok := v.([]byte); ok {
			// destination byte slice has to be of correct size
			e := make([]byte, len(b))
			d[i] = &e
		} else if b, ok = v.(ByteSlice); ok {
			// destination ByteSlice has to be of correct size
			e := make(ByteSlice, len(b))
			d[i] = &e
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
