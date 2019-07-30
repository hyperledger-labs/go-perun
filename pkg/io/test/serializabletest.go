// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package test contains the generic serializable tests.
package test // import "perun.network/go-perun/pkg/io/test"

import (
	"io"
	"reflect"
	"testing"

	perunio "perun.network/go-perun/pkg/io"
)

// GenericSerializableTest runs multiple tests to check whether encoding
// and decoding of serializable values works.
func GenericSerializableTest(t *testing.T, serializables ...perunio.Serializable) {
	genericDecodeEncodeTest(t, serializables...)
	genericBrokenPipeTests(t, serializables...)
}

// genericDecodeEncodeTest tests whether encoding and then decoding
// serializable values results in the original values.
func genericDecodeEncodeTest(t *testing.T, serializables ...perunio.Serializable) {
	r, w := io.Pipe()
	done := make(chan struct{})
	go func(serializables []perunio.Serializable) {

		for i, v := range serializables {
			if err := perunio.Encode(w, v); err != nil {
				t.Errorf("failed to encode %dth element (%T): %+v", i, v, err)
			}
		}
		close(done)
	}(serializables)

	for i, v := range serializables {

		dest := reflect.New(reflect.TypeOf(v).Elem())

		if err := perunio.Decode(r, dest.Interface().(perunio.Serializable)); err != nil {
			t.Errorf("failed to decode %dth element (%T): %+v", i, v, err)
		}

		if !reflect.DeepEqual(v, dest.Interface()) {
			t.Errorf("encoding and decoding the %dth element (%T) resulted in different value: %v, %v", i, v, reflect.ValueOf(v).Elem(), dest.Elem())
		}
	}
	<-done
}

func genericBrokenPipeTests(t *testing.T, serializables ...perunio.Serializable) {
	for i, v := range serializables {
		r, w := io.Pipe()
		_ = w.Close()
		if err := v.Encode(w); err == nil {
			t.Errorf("encoding on closed writer should fail, but does not. %dth element (%T)", i, v)
		}

		_ = r.Close()
		if err := v.Decode(r); err == nil {
			t.Errorf("encoding on closed reader should fail, but does not. %dth element (%T)", i, v)
		}
	}
}
