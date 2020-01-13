// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

// Package test contains the generic serializer tests.
package test

import (
	"io"
	"reflect"
	"testing"
	"testing/iotest"

	perunio "perun.network/go-perun/pkg/io"
)

// GenericSerializerTest runs multiple tests to check whether encoding
// and decoding of serializer values works.
func GenericSerializerTest(t *testing.T, serializers ...perunio.Serializer) {
	genericDecodeEncodeTest(t, serializers...)
	GenericBrokenPipeTest(t, serializers...)
}

// genericDecodeEncodeTest tests whether encoding and then decoding
// serializer values results in the original values.
func genericDecodeEncodeTest(t *testing.T, serializers ...perunio.Serializer) {
	for i, v := range serializers {
		r, w := io.Pipe()
		br := iotest.OneByteReader(r)
		go func() {
			if err := perunio.Encode(w, v); err != nil {
				t.Errorf("failed to encode %dth element (%T): %+v", i, v, err)
			}
			w.Close()
		}()

		dest := reflect.New(reflect.TypeOf(v).Elem())
		err := perunio.Decode(br, dest.Interface().(perunio.Serializer))
		r.Close()
		if err != nil {
			t.Errorf("failed to decode %dth element (%T): %+v", i, v, err)
		} else if !reflect.DeepEqual(v, dest.Interface()) {
			t.Errorf(
				"encoding and decoding the %dth element (%T) resulted in different value: %v, %v",
				i, v, reflect.ValueOf(v).Elem(), dest.Elem())
		}
	}
}

// GenericBrokenPipeTest tests that encoding and decoding on broken streams fails.
func GenericBrokenPipeTest(t *testing.T, serializers ...perunio.Serializer) {
	for i, v := range serializers {
		r, w := io.Pipe()
		_ = w.Close()
		if err := v.Encode(w); err == nil {
			t.Errorf("encoding on closed writer should fail, but does not. %dth element (%T)", i, v)
		}

		_ = r.Close()
		if err := v.Decode(r); err == nil {
			t.Errorf("decoding on closed reader should fail, but does not. %dth element (%T)", i, v)
		}
	}
}
