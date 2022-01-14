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

package wire_test

import (
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/test"
)

var nilDecoder = func(io.Reader) (wire.Msg, error) { return nil, nil }

func TestType_Valid_String(t *testing.T) {
	test.OnlyOnce(t)

	const testTypeVal, testTypeStr = 252, "testTypeA"
	testType := wire.Type(testTypeVal)
	assert.False(t, testType.Valid(), "unregistered type should not be valid")
	assert.Equal(t, strconv.Itoa(testTypeVal), testType.String(),
		"unregistered type's String() should return its integer value")

	wire.RegisterExternalDecoder(testTypeVal, nilDecoder, testTypeStr)
	assert.True(t, testType.Valid(), "registered type should be valid")
	assert.Equal(t, testTypeStr, testType.String(),
		"registered type's String() should be 'testType'")
}

func TestRegisterExternalDecoder(t *testing.T) {
	test.OnlyOnce(t)

	const testTypeVal, testTypeStr = 251, "testTypeB"

	wire.RegisterExternalDecoder(testTypeVal, nilDecoder, testTypeStr)
	assert.Panics(t,
		func() { wire.RegisterExternalDecoder(testTypeVal, nilDecoder, testTypeStr) },
		"second registration of same type should fail",
	)
	assert.Panics(t,
		func() { wire.RegisterExternalDecoder(wire.Ping, nilDecoder, "PingFail") },
		"registration of internal type should fail",
	)
}
