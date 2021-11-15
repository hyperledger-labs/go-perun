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

package wire

import (
	"io"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	wtest "perun.network/go-perun/wallet/test"
	iotest "polycry.pt/poly-go/io/test"
	"polycry.pt/poly-go/test"
)

// NewRandomEnvelope - copy from wire/test for internal tests.
func NewRandomEnvelope(rng *rand.Rand, m Msg) *Envelope {
	return &Envelope{
		Sender:    wtest.NewRandomAddress(rng),
		Recipient: wtest.NewRandomAddress(rng),
		Msg:       m,
	}
}

var nilDecoder = func(io.Reader) (Msg, error) { return nil, nil }

func TestType_Valid_String(t *testing.T) {
	test.OnlyOnce(t)

	const testTypeVal, testTypeStr = 252, "testTypeA"
	testType := Type(testTypeVal)
	assert.False(t, testType.Valid(), "unregistered type should not be valid")
	assert.Equal(t, strconv.Itoa(testTypeVal), testType.String(),
		"unregistered type's String() should return its integer value")

	RegisterExternalDecoder(testTypeVal, nilDecoder, testTypeStr)
	assert.True(t, testType.Valid(), "registered type should be valid")
	assert.Equal(t, testTypeStr, testType.String(),
		"registered type's String() should be 'testType'")
}

func TestRegisterExternalDecoder(t *testing.T) {
	test.OnlyOnce(t)

	const testTypeVal, testTypeStr = 251, "testTypeB"

	RegisterExternalDecoder(testTypeVal, nilDecoder, testTypeStr)
	assert.Panics(t,
		func() { RegisterExternalDecoder(testTypeVal, nilDecoder, testTypeStr) },
		"second registration of same type should fail",
	)
	assert.Panics(t,
		func() { RegisterExternalDecoder(Ping, nilDecoder, "PingFail") },
		"registration of internal type should fail",
	)
}

func TestEnvelope_EncodeDecode(t *testing.T) {
	ping := NewRandomEnvelope(test.Prng(t), NewPingMsg())
	iotest.GenericSerializerTest(t, ping)
}
