// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package io_test

import (
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	perunio "perun.network/go-perun/pkg/io"
	iotest "perun.network/go-perun/pkg/io/test"
)

func TestByteSlice(t *testing.T) {
	var v1, v2, v3, v4 perunio.ByteSlice = []byte{}, []byte{255}, []byte{1, 2, 3, 4, 5, 6}, make([]byte, 20000)
	testByteSlices(t, v1, v2, v3, v4)
	iotest.GenericBrokenPipeTest(t, &v1, &v2, &v3, &v4)
}

// TestStutter tests what happens if the network stutters (split one message into several network packages).
func TestStutter(t *testing.T) {
	var values = []byte{0, 1, 2, 3, 4, 5, 6, 255}
	r, w := io.Pipe()

	go func() {
		for _, v := range values {
			w.Write([]byte{v})
		}
	}()

	var decodedValue perunio.ByteSlice = make([]byte, len(values))
	assert.Nil(t, decodedValue.Decode(r))
	for i, v := range values {
		assert.Equal(t, decodedValue[i], v)
	}

}

func testByteSlices(t *testing.T, serial ...perunio.ByteSlice) {
	a := assert.New(t)
	r, w := io.Pipe()
	done := make(chan struct{})

	go func() {
		for _, v := range serial {
			a.NoError(v.Encode(w))
		}
		close(done)
	}()

	for i, v := range serial {

		d := make([]byte, len(v))
		dest := perunio.ByteSlice(d)

		a.NoError(dest.Decode(r), "failed to decode element")

		if !reflect.DeepEqual(v, dest) {
			t.Errorf("encoding and decoding the %dth element (%T) resulted in different value: %v, %v", i, v, reflect.ValueOf(v).Elem(), dest)
		}
	}
	<-done
}
