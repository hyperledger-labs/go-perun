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

package channel_test

import (
	"testing"

	_ "perun.network/go-perun/backend/sim" // backend init
	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	iotest "perun.network/go-perun/pkg/io/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func TestStateSerialization(t *testing.T) {
	rng := pkgtest.Prng(t)
	state := test.NewRandomState(rng, test.WithNumLocked(int(rng.Int31n(4)+1)))
	state2 := test.NewRandomState(rng, test.WithIsFinal(!state.IsFinal), test.WithNumLocked(int(rng.Int31n(4)+1)))

	iotest.GenericSerializerTest(t, state)
	test.GenericStateEqualTest(t, state, state2)

	state.App = channel.NoApp()
	state.Data = channel.NoData()
	iotest.GenericSerializerTest(t, state)
}
