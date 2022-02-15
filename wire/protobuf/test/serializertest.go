// Copyright 2022 - See NOTICE file for copyright holders.
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
	"bytes"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
	pkgtest "polycry.pt/poly-go/test"
)

// MsgSerializerTest runs multiple tests to check whether encoding and decoding
// of msg values works.
func MsgSerializerTest(t *testing.T, msg wire.Msg) {
	t.Helper()

	rng := pkgtest.Prng(t)
	envelope := newEnvelope(rng)
	envelope.Msg = msg

	var buff bytes.Buffer
	require.NoError(t, wire.EncodeEnvelope(&buff, envelope))

	gotEnvelope, err := wire.DecodeEnvelope(&buff)
	require.NoError(t, err)
	assert.EqualValues(t, envelope, gotEnvelope)
}

func newEnvelope(rng *rand.Rand) *wire.Envelope {
	return &wire.Envelope{
		Sender:    wallettest.NewRandomAddress(rng),
		Recipient: wallettest.NewRandomAddress(rng),
	}
}
