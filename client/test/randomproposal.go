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

package test

import (
	"math/rand"

	channeltest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	wiretest "perun.network/go-perun/wire/test"
)

// NewRandomLedgerChannelProposal creates a random channel proposal with the supplied
// options. Number of participants is fixed to 2.
func NewRandomLedgerChannelProposal(rng *rand.Rand, opts ...client.ProposalOpts) *client.LedgerChannelProposal {
	return NewRandomLedgerChannelProposalBy(rng, wallettest.NewRandomAddress(rng), opts...)
}

// NewRandomLedgerChannelProposalBy creates a random channel proposal with the
// supplied options and proposer. Number of participants is fixed to 2.
func NewRandomLedgerChannelProposalBy(rng *rand.Rand, proposer wallet.Address, opts ...client.ProposalOpts) *client.LedgerChannelProposal {
	return client.NewLedgerChannelProposal(
		rng.Uint64(),
		proposer,
		channeltest.NewRandomAllocation(rng, channeltest.WithNumParts(2)),
		wiretest.NewRandomAddresses(rng, 2),
		opts...)
}

// NewRandomSubChannelProposal creates a random subchannel proposal with the
// supplied options. Number of participants is fixed to 2.
func NewRandomSubChannelProposal(rng *rand.Rand, opts ...client.ProposalOpts) *client.SubChannelProposal {
	return client.NewSubChannelProposal(
		channeltest.NewRandomChannelID(rng),
		rng.Uint64(),
		channeltest.NewRandomAllocation(rng, channeltest.WithNumParts(2)),
		opts...)
}
