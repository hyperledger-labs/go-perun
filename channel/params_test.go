// Copyright 2024 - See NOTICE file for copyright holders.
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

package channel_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/wire/perunio"
	peruniotest "perun.network/go-perun/wire/perunio/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestParams_Clone(t *testing.T) {
	rng := pkgtest.Prng(t)
	params := test.NewRandomParams(rng)
	clone := params.Clone()

	require.Equalf(t, params.Parts, clone.Parts, "Clone() = %v, want %v", clone, params)
	require.Equalf(t, params.App, clone.App, "Clone() = %v, want %v", clone, params)
	require.Equalf(t, params.ChallengeDuration, clone.ChallengeDuration, "Clone() = %v, want %v", clone, params)
	require.Equalf(t, params.Nonce, clone.Nonce, "Clone() = %v, want %v", clone, params)
}

func TestParams_Serializer(t *testing.T) {
	rng := pkgtest.Prng(t)
	params := make([]perunio.Serializer, 10)
	for i := range params {
		var p *channel.Params
		if i&1 == 0 {
			p = test.NewRandomParams(rng, test.WithoutApp())
		} else {
			p = test.NewRandomParams(rng)
		}
		params[i] = p
	}

	peruniotest.GenericSerializerTest(t, params...)
}
