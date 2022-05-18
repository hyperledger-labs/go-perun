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

package client_test

import (
	"context"
	"math/big"
	"testing"

	"perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
	pkgtest "polycry.pt/poly-go/test"
)

func TestProgression(t *testing.T) {
	rng := pkgtest.Prng(t)

	setups, errs := NewSetups(rng, []string{"Paul", "Paula"})
	roles := [2]clienttest.Executer{
		clienttest.NewPaul(t, setups[0]),
		clienttest.NewPaula(t, setups[1]),
	}

	appAddress := test.NewRandomAddress(rng)
	app := channel.NewMockApp(appAddress)
	channel.RegisterApp(app)

	execConfig := &clienttest.ProgressionExecConfig{
		BaseExecConfig: clienttest.MakeBaseExecConfig(
			[2]wire.Address{setups[0].Identity.Address(), setups[1].Identity.Address()},
			chtest.NewRandomAsset(rng),
			[2]*big.Int{big.NewInt(99), big.NewInt(1)},
			client.WithApp(app, channel.NewMockOp(channel.OpValid)),
		),
	}

	ctx, cancel := context.WithTimeout(context.Background(), twoPartyTestTimeout)
	defer cancel()
	clienttest.ExecuteTwoPartyTest(ctx, t, roles, execConfig, errs)
}
