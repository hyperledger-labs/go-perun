// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package test contains the generic serializable tests.
package test // import "perun.network/go-perun/pkg/io/test"

import (
	_io "io"
	"reflect"
	"testing"

	"perun.network/go-perun/pkg/io"
)

// GenericSerializableTest runs multiple tests to check whether encoding
// and decoding of serializable values works.
func GenericSerializableTest(t *testing.T, serializables ...io.Serializable) {
	genericDecodeEncodeTest(t, serializables...)
	genericBrokenPipeTests(t, serializables...)
}

// genericDecodeEncodeTest tests whether encoding and then decoding
// serializable values results in the original values.
func genericDecodeEncodeTest(t *testing.T, serializables ...io.Serializable) {
	for i, v := range serializables {
		r, w := _io.Pipe()

		dest := reflect.New(reflect.TypeOf(v).Elem())

		go func(i int, v io.Serializable) {
			if err := io.Encode(w, v); err != nil {
				t.Errorf("failed to encode %dth element (%T): %+v", i, v, err)
			}
		}(i, v)

		if err := io.Decode(r, dest.Interface().(io.Serializable)); err != nil {
			t.Errorf("failed to decode %dth element (%T): %+v", i, v, err)
		}

		if !reflect.DeepEqual(v, dest.Interface()) {
			t.Errorf("encoding and decoding the %dth element (%T) resulted in different value: %v, %v", i, v, reflect.ValueOf(v).Elem(), dest.Elem())
		}
	}
}

func genericBrokenPipeTests(t *testing.T, serializables ...io.Serializable) {
	for i, v := range serializables {
		r, w := _io.Pipe()
		w.Close()
		if err := v.Encode(w); err == nil {
			t.Errorf("encoding on closed writer should fail, but does not. %dth element (%T)", i, v)
		}

		r.Close()
		if err := v.Decode(r); err == nil {
			t.Errorf("encoding on closed writer should fail, but does not. %dth element (%T)", i, v)
		}
	}
}
