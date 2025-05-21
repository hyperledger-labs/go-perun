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

package wire_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/wire"
	wiretest "perun.network/go-perun/wire/test"
	"polycry.pt/poly-go/test"
)

func TestProducer_produce_closed(t *testing.T) {
	var missed *wire.Envelope
	p := wire.NewRelay()
	p.SetDefaultMsgHandler(func(e *wire.Envelope) { missed = e })
	require.NoError(t, p.Close())
	rng := test.Prng(t)
	a := wiretest.NewRandomAddress(rng)
	b := wiretest.NewRandomAddress(rng)
	p.Put(&wire.Envelope{a, b, wire.NewPingMsg()})
	assert.Nil(t, missed, "produce() on closed producer shouldn't do anything")
}
