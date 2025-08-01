// Copyright 2021 - See NOTICE file for copyright holders.
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
	"encoding"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type binary interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// GenericMarshalerTest runs multiple tests to check whether encoding
// and decoding of serializer values works.
func GenericMarshalerTest(t *testing.T, serializers ...binary) {
	t.Helper()

	for i, v := range serializers {
		data, err := v.MarshalBinary()
		require.NoError(t, err, "failed to encode %dth element (%T)", i, v)

		dest := reflect.New(reflect.TypeOf(v).Elem()).
			Interface().(encoding.BinaryUnmarshaler)
		err = dest.UnmarshalBinary(data)
		require.NoError(t, err, "failed to decode %dth element (%T)", i, v)

		assert.Equalf(t, v, dest, "comparing element %d", i)
	}
}
