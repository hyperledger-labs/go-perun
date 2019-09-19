// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"io"
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/io/test"
	peruntest "perun.network/go-perun/pkg/test"
)

func TestBigInt(t *testing.T) {
	var v1 = BigInt{big.NewInt(123456)}
	var v2 = BigInt{big.NewInt(1)}
	var v3 = BigInt{big.NewInt(0)}
	test.GenericSerializableTest(t, &v1, &v2, &v3)
}

func TestInvalidBigInt(t *testing.T) {
func TestBigInt_Negative(t *testing.T) {
	neg, buf := BigInt{big.NewInt(-1)}, new(bytes.Buffer)
	assert.Panics(t, func() { neg.Encode(buf) }, "encoding negative big.Int should panic")
	assert.Zero(t, buf.Len(), "encoding negative big.Int should not write anything")
}

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

	_, w = io.Pipe()

	if panics, _ := peruntest.CheckPanic(func() { BigInt{nil}.Encode(w) }); !panics {
		t.Error("encoding nil BigInt failed to panic")
	}
}
