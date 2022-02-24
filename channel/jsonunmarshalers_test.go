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

package channel_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"

	"polycry.pt/poly-go/test"
)

func TestStateJSONMarshaling(t *testing.T) {
	rng := test.Prng(t)
	state := chtest.NewRandomState(rng, chtest.WithoutApp())

	b, err := json.Marshal(state)
	require.NoError(t, err)

	state1 := new(channel.State)
	require.NoError(t, json.Unmarshal(b, state1))
	require.Equal(t, state, state1)
}

func TestParamsJSONMarshaling(t *testing.T) {
	rng := test.Prng(t)
	params := chtest.NewRandomParams(rng, chtest.WithoutApp())

	b, err := json.Marshal(params)
	require.NoError(t, err)

	params1 := new(channel.Params)
	require.NoError(t, json.Unmarshal(b, params1))
	require.Equal(t, params, params1)
}
