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
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/pkg/io/test"
)

func TestBigInt_Generic(t *testing.T) {
	vars := []perunio.Serializer{
		&perunio.BigInt{big.NewInt(0)},
		&perunio.BigInt{big.NewInt(1)},
		&perunio.BigInt{big.NewInt(123456)},
		&perunio.BigInt{new(big.Int).SetBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})}, // larger than uint64
	}
	test.GenericSerializerTest(t, vars...)
}

func TestBigInt_DecodeZeroLength(t *testing.T) {
	assert := assert.New(t)

	buf := bytes.NewBuffer([]byte{0})
	var result perunio.BigInt
	assert.NoError(result.Decode(buf), "decoding zero length big.Int should work")
	assert.Zero(new(big.Int).Cmp(result.Int), "decoding zero length should set big.Int to 0")
}

func TestBigInt_DecodeToExisting(t *testing.T) {
	x, buf := new(big.Int), bytes.NewBuffer([]byte{1, 42})
	wx := perunio.BigInt{x}
	assert.NoError(t, wx.Decode(buf), "decoding {1, 42} into big.Int should work")
	assert.Zero(t, big.NewInt(42).Cmp(x), "decoding {1, 42} into big.Int should result in 42")
}

func TestBigInt_Negative(t *testing.T) {
	neg, buf := perunio.BigInt{big.NewInt(-1)}, new(bytes.Buffer)
	assert.Panics(t, func() { neg.Encode(buf) }, "encoding negative big.Int should panic")
	assert.Zero(t, buf.Len(), "encoding negative big.Int should not write anything")
}

func TestBigInt_Invalid(t *testing.T) {
	a := assert.New(t)
	buf := new(bytes.Buffer)
	// Test integers that are too big
	tooBigBitPos := []uint{perunio.MaxBigIntLength*8 + 1, 0xff*8 + 1} // too big uint8 and uint16 lengths
	for _, pos := range tooBigBitPos {
		var tooBig = perunio.BigInt{big.NewInt(1)}
		tooBig.Lsh(tooBig.Int, pos)

		a.Error(tooBig.Encode(buf), "encoding too big big.Int should fail")
		a.Zero(buf.Len(), "encoding too big big.Int should not have written anything")
		buf.Reset() // in case above test failed
	}

	// manually encode too big number to test failing of decoding
	buf.Write([]byte{perunio.MaxBigIntLength + 1})
	for i := 0; i < perunio.MaxBigIntLength+1; i++ {
		buf.WriteByte(0xff)
	}

	var result perunio.BigInt
	a.Error(result.Decode(buf), "decoding of an integer that is too big should fail")
	buf.Reset()

	// Test not sending value, only length
	buf.WriteByte(1)
	a.Error(result.Decode(buf), "decoding after sender only sent length should fail")

	a.Panics(func() { perunio.BigInt{nil}.Encode(buf) }, "encoding nil big.Int failed to panic")
}
