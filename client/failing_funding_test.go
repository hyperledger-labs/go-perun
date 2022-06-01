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
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/test"
)

func TestFailingFunding(t *testing.T) {
	const (
		fundTimeout  = 1 * time.Second
		fridaIdx     = 0
		fredIdx      = 1
		fridaInitBal = 100
		fredInitBal  = 50
	)

	ctx, cancel := context.WithTimeout(context.Background(), twoPartyTestTimeout)
	defer cancel()
	// The ctxFundTimeout is used for Fred when accepting the proposal. This
	// context deadline will exceed since Frida's funding fails.
	ctxFundTimeout, cancelFundCtx := context.WithTimeout(context.Background(), fundTimeout)
	defer cancelFundCtx()
	rng := test.Prng(t)

	setups := NewSetups(rng, []string{"Frida", "Fred"})
	// Inject the failing funder for Frida.
	setups[fridaIdx].Funder = FailingFunder{}
	clients := NewClients(t, rng, setups)
	frida, fred := clients[fridaIdx], clients[fredIdx]
	fridaAddr, fredAddr := frida.Identity.Address(), fred.Identity.Address()

	// The channel into which Fred's created ledger channel is sent into.
	chsFred := make(chan *client.Channel, 1)
	// The channel handlers for Fred.
	var chHandlerFred client.ProposalHandlerFunc = func(cp client.ChannelProposal, pr *client.ProposalResponder) {
		switch cp := cp.(type) {
		case *client.LedgerChannelProposalMsg:
			ch, err := pr.Accept(ctxFundTimeout, cp.Accept(fredAddr, client.WithRandomNonce()))
			require.Error(t, err)
			chsFred <- ch
		default:
			t.Fatalf("expected ledger channel proposal, got %T", cp)
		}
	}
	var chUpHandlerFred client.UpdateHandlerFunc

	go fred.Handle(chHandlerFred, chUpHandlerFred)

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
	chFrida, err := frida.ProposeChannel(ctx, prop)
	require.Error(t, err)
	require.NotNil(t, chFrida)
	// Frida settles the channel.
	require.NoError(t, chFrida.Settle(ctx, false))

	// Fred gets the channel and settles it afterwards.
	chFred := <-chsFred
	require.NotNil(t, chFred)
	// Fred settles the channel as a secondary.
	require.NoError(t, chFred.Settle(ctx, true))

	// Test the final balances.
	fridaFinalBal := frida.BalanceReader.Balance(fridaAddr, asset)
	assert.Truef(t, fridaFinalBal.Cmp(big.NewInt(fridaInitBal)) == 0, "frida: wrong final balance: got %v, expected %v", fridaFinalBal, fridaInitBal)
	fredFinalBal := fred.BalanceReader.Balance(fredAddr, asset)
	assert.Truef(t, fredFinalBal.Cmp(big.NewInt(fredInitBal)) == 0, "fred: wrong final balance: got %v, expected %v", fredFinalBal, fredInitBal)
}

type FailingFunder struct{}

// Fund returns an error to simulate failed funding.
func (m FailingFunder) Fund(ctx context.Context, req channel.FundingReq) error {
	return errors.New("funding failed")
}
