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

package test

import (
	"io"
	"reflect"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	perunio "perun.network/go-perun/pkg/io"
)

// GenericSerializerTest runs multiple tests to check whether encoding
// and decoding of serializer values works.
func GenericSerializerTest(t *testing.T, serializers ...perunio.Serializer) {
	t.Helper()
	genericDecodeEncodeTest(t, serializers...)
	GenericBrokenPipeTest(t, serializers...)
}

// genericDecodeEncodeTest tests whether encoding and then decoding
// serializer values results in the original values.
func genericDecodeEncodeTest(t *testing.T, serializers ...perunio.Serializer) {
	t.Helper()
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
		} else {
			_v := dest.Interface()
			assert.Equalf(t, v, _v, "comparing element %d", i)
		}
	}
}

// GenericBrokenPipeTest tests that encoding and decoding on broken streams fails.
func GenericBrokenPipeTest(t *testing.T, serializers ...perunio.Serializer) {
	t.Helper()
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
