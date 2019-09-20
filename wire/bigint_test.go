// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/io"
	"perun.network/go-perun/pkg/io/test"
)

func TestBigInt_Generic(t *testing.T) {
	vars := []io.Serializable{
		&BigInt{big.NewInt(0)},
		&BigInt{big.NewInt(1)},
		&BigInt{big.NewInt(123456)},
		&BigInt{new(big.Int).SetBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})}, // larger than uint64
	}
	test.GenericSerializableTest(t, vars...)
}

func TestBigInt_DecodeZeroLength(t *testing.T) {
	assert := assert.New(t)

	buf := bytes.NewBuffer([]byte{0})
	var result BigInt
	assert.NoError(result.Decode(buf), "decoding zero length big.Int should work")
	assert.Zero(new(big.Int).Cmp(result.Int), "decoding zero length should set big.Int to 0")
}

func TestBigInt_DecodeToExisting(t *testing.T) {
	x, buf := new(big.Int), bytes.NewBuffer([]byte{1, 42})
	wx := BigInt{x}
	assert.NoError(t, wx.Decode(buf), "decoding {1, 42} into big.Int should work")
	assert.Zero(t, big.NewInt(42).Cmp(x), "decoding {1, 42} into big.Int should result in 42")
}

func TestBigInt_Negative(t *testing.T) {
	neg, buf := BigInt{big.NewInt(-1)}, new(bytes.Buffer)
	assert.Panics(t, func() { neg.Encode(buf) }, "encoding negative big.Int should panic")
	assert.Zero(t, buf.Len(), "encoding negative big.Int should not write anything")
}

func TestBigInt_Invalid(t *testing.T) {
	a := assert.New(t)
	buf := new(bytes.Buffer)
	// Test integers that are too big
	tooBigBitPos := []uint{maxBigIntLength*8 + 1, 0xff*8 + 1} // too big uint8 and uint16 lengths
	for _, pos := range tooBigBitPos {
		var tooBig = BigInt{big.NewInt(1)}
		tooBig.Lsh(tooBig.Int, pos)

		a.Error(tooBig.Encode(buf), "encoding too big big.Int should fail")
		a.Zero(buf.Len(), "encoding too big big.Int should not have written anything")
		buf.Reset() // in case above test failed
	}

	// manually encode too big number to test failing of decoding
	buf.Write([]byte{maxBigIntLength + 1})
	for i := 0; i < maxBigIntLength+1; i++ {
		buf.WriteByte(0xff)
	}

	var result BigInt
	a.Error(result.Decode(buf), "decoding of an integer that is too big should fail")
	buf.Reset()

	// Test not sending value, only length
	buf.WriteByte(1)
	a.Error(result.Decode(buf), "decoding after sender only sent length should fail")

	a.Panics(func() { BigInt{nil}.Encode(buf) }, "encoding nil big.Int failed to panic")
}
