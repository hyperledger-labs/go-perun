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

package channel

import (
	"testing"

	chtest "perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wallet"
	wtest "perun.network/go-perun/wallet/test"
)

func TestGenericTests(t *testing.T) {
	setup := newChannelSetup(t)
	chtest.GenericBackendTest(t, setup)
}

func newChannelSetup(t *testing.T) *chtest.Setup {
	rng := pkgtest.Prng(t)

	params, state := chtest.NewRandomParamsAndState(rng, chtest.WithNumLocked(int(rng.Int31n(4)+1)))
	params2, state2 := chtest.NewRandomParamsAndState(rng, chtest.WithIsFinal(!state.IsFinal), chtest.WithNumLocked(int(rng.Int31n(4)+1)))

	return &chtest.Setup{
		Params:        params,
		Params2:       params2,
		State:         state,
		State2:        state2,
		Account:       wtest.NewRandomAccount(rng),
		RandomAddress: func() wallet.Address { return wtest.NewRandomAddress(rng) },
	}
}
