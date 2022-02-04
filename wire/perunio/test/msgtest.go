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
	"math/rand"
	"testing"

	"perun.network/go-perun/wire"
	perunioserializer "perun.network/go-perun/wire/perunio/serializer"

	wiretest "perun.network/go-perun/wire/test"
	pkgtest "polycry.pt/poly-go/test"
)

// serializableEnvelope implements perunio serializer for wire.Envelope, so
// that generic serialzer tests can be run for envelopes.
type serializableEnvelope struct {
	env *wire.Envelope
}

func (e *serializableEnvelope) Encode(writer io.Writer) error {
	return (perunioserializer.Serializer{}).Encode(writer, e.env)
}

func (e *serializableEnvelope) Decode(reader io.Reader) (err error) {
	e.env, err = (perunioserializer.Serializer{}).Decode(reader)
	return err
}

func newSerializableEnvelope(rng *rand.Rand, msg wire.Msg) *serializableEnvelope {
	return &serializableEnvelope{env: wiretest.NewRandomEnvelope(rng, msg)}
}

// MsgSerializerTest performs generic serializer tests on a wire.Msg object.
// It tests the perunio encoder/decoder implementations on the individual types
// and the registration of the corresponding decoders.
func MsgSerializerTest(t *testing.T, msg wire.Msg) {
	t.Helper()

	GenericSerializerTest(t, newSerializableEnvelope(pkgtest.Prng(t), msg))
}
