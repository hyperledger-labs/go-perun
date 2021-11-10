// Copyright 2020 - See NOTICE file for copyright holders.
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
	wtest "perun.network/go-perun/wallet/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestMachineClone(t *testing.T) {
	rng := pkgtest.Prng(t)

	acc := wtest.NewRandomAccount(rng)
	params := *test.NewRandomParams(rng, test.WithFirstPart(acc.Address()))

	sm, err := channel.NewStateMachine(acc, params)
	require.NoError(t, err)
	pkgtest.VerifyClone(t, sm)

	am, err := channel.NewActionMachine(acc, params)
	require.NoError(t, err)
	pkgtest.VerifyClone(t, am)
}
