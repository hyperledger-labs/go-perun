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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	clienttest "perun.network/go-perun/client/test"
	_ "perun.network/go-perun/wire/perunio/serializer" // wire serialzer init
	peruniotest "perun.network/go-perun/wire/perunio/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestProposalMsgsSerialization(t *testing.T) {
	clienttest.ProposalMsgsSerializationTest(t, peruniotest.MsgSerializerTest)
}

func TestNewLedgerChannelProposal(t *testing.T) {
	rng := pkgtest.Prng(t)
	base := clienttest.NewRandomLedgerChannelProposal(rng)

	// Without FundingAgreement it uses InitBals.
	prop, err := client.NewLedgerChannelProposal(base.ChallengeDuration, base.Participant, base.InitBals, base.Peers)
	require.NoError(t, err)
	assert.Equal(t, base.InitBals.Balances, prop.FundingAgreement)

	// FundingAgreements number of assets do not match InitBals.
	agreement := test.NewRandomBalances(rng, test.WithNumAssets(len(base.InitBals.Assets)+1))
	_, err = client.NewLedgerChannelProposal(base.ChallengeDuration, base.Participant, base.InitBals, base.Peers, client.WithFundingAgreement(agreement))
	assert.EqualError(t, err, "comparing FundingAgreement and initial balances sum: dimension mismatch")

	// FundingAgreements sum do not match InitBals sum.
	agreement = test.NewRandomBalances(rng, test.WithNumAssets(len(base.InitBals.Assets)))
	_, err = client.NewLedgerChannelProposal(base.ChallengeDuration, base.Participant, base.InitBals, base.Peers, client.WithFundingAgreement(agreement))
	assert.EqualError(t, err, "FundingAgreement and initial balances differ")
}

func TestLedgerChannelProposalReqProposalID(t *testing.T) {
	rng := pkgtest.Prng(t)
	original := *client.NewRandomLedgerChannelProposal(rng)

	fake := *client.NewRandomLedgerChannelProposal(rng)

	assert.NotEqual(t, original.ProposalID, fake.ProposalID)
}

func TestSubChannelProposalReqProposalID(t *testing.T) {
	rng := pkgtest.Prng(t)
	original, err := clienttest.NewRandomSubChannelProposal(rng)
	require.NoError(t, err)

	fake, err := clienttest.NewRandomSubChannelProposal(rng)
	require.NoError(t, err)

	assert.NotEqual(t, original.ProposalID, fake.ProposalID)
}

func TestVirtualChannelProposalReqProposalID(t *testing.T) {
	rng := pkgtest.Prng(t)

	original, err := clienttest.NewRandomVirtualChannelProposal(rng)
	require.NoError(t, err)
	assert.Equal(t, channel.NoApp(), original.App, "virtual channel should always has no app")
	assert.Equal(t, channel.NoData(), original.InitData, "virtual channel should always has no data")

	fake, err := clienttest.NewRandomVirtualChannelProposal(rng)
	require.NoError(t, err)
	assert.Equal(t, channel.NoApp(), fake.App, "virtual channel should always has no app")
	assert.Equal(t, channel.NoData(), fake.InitData, "virtual channel should always has no data")

	assert.NotEqual(t, original.ProposalID, fake.ProposalID)
}
