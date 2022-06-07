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
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	ctest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/test"
)

func TestFailingFunding(t *testing.T) {
	rng := test.Prng(t)

	t.Run("failing funder proposer", func(t *testing.T) {
		setups := NewSetups(rng, []string{"Frida", "Fred"})
		setups[0].Funder = FailingFunder{}

		runFredFridaTest(t, rng, setups)
	})

	t.Run("failing funder proposee", func(t *testing.T) {
		setups := NewSetups(rng, []string{"Frida", "Fred"})
		setups[1].Funder = FailingFunder{}

		runFredFridaTest(t, rng, setups)
	})

	t.Run("failing funder both sides", func(t *testing.T) {
		setups := NewSetups(rng, []string{"Frida", "Fred"})
		setups[0].Funder = FailingFunder{}
		setups[1].Funder = FailingFunder{}

		runFredFridaTest(t, rng, setups)
	})
}

func runFredFridaTest(t *testing.T, rng *rand.Rand, setups []ctest.RoleSetup) {
	t.Helper()
	const (
		challengeDuration = 1
		fridaIdx          = 0
		fredIdx           = 1
		fridaInitBal      = 100
		fredInitBal       = 50
	)

	ctx, cancel := context.WithTimeout(context.Background(), twoPartyTestTimeout)
	defer cancel()

	clients := NewClients(t, rng, setups)
	frida, fred := clients[fridaIdx], clients[fredIdx]
	fridaAddr, fredAddr := frida.Identity.Address(), fred.Identity.Address()

	chsFred := make(chan *client.Channel, 1)
	errsFred := make(chan error)
	go fred.Handle(
		ctest.AlwaysAcceptChannelHandler(ctx, fredAddr, chsFred, errsFred),
		ctest.AlwaysRejectUpdateHandler(ctx, errsFred),
	)

	// Create the proposal.
	asset := chtest.NewRandomAsset(rng)
	initAlloc := channel.NewAllocation(2, asset)
	initAlloc.SetAssetBalances(asset, []*big.Int{big.NewInt(fridaInitBal), big.NewInt(fredInitBal)})
	parts := []wire.Address{fridaAddr, fredAddr}
	prop, err := client.NewLedgerChannelProposal(
		challengeDuration,
		fridaAddr,
		initAlloc,
		parts,
	)
	require.NoError(t, err, "creating ledger channel proposal")

	// Frida sends the proposal.
	_, err = frida.ProposeChannel(ctx, prop)
	checkChannelFundingError(t, err)

	select {
	case <-chsFred:
		t.Fatalf("expected channel creation to fail")
	case err = <-errsFred:
	}
	checkChannelFundingError(t, err)

	// Test the final balances.
	fridaFinalBal := frida.BalanceReader.Balance(fridaAddr, asset)
	assert.Truef(t, fridaFinalBal.Cmp(big.NewInt(fridaInitBal)) == 0, "frida: wrong final balance: got %v, expected %v", fridaFinalBal, fridaInitBal)
	fredFinalBal := fred.BalanceReader.Balance(fredAddr, asset)
	assert.Truef(t, fredFinalBal.Cmp(big.NewInt(fredInitBal)) == 0, "fred: wrong final balance: got %v, expected %v", fredFinalBal, fredInitBal)
}

// checkChannelFundingError checks whether the given error is of type
// ChannelFundingError and the settling hasn't failed.
func checkChannelFundingError(t *testing.T, err error) {
	t.Helper()
	var cfError client.ChannelFundingError
	require.Truef(t, errors.As(err, &cfError), fmt.Sprintf("expexted ChannelFundingError, got %T", err))
	require.Nil(t, cfError.SettleError)
}

type FailingFunder struct{}

// Fund returns an error to simulate failed funding.
func (m FailingFunder) Fund(ctx context.Context, req channel.FundingReq) error {
	return errors.New("funding failed")
}
