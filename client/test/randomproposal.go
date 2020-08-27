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
)

// NewRandomChannelProposal creates a random channel proposal with the supplied
// options. Number of participants is fixed to 2.
func NewRandomChannelProposal(rng *rand.Rand, opts ...client.ProposalOpts) client.ChannelProposal {
	return NewRandomChannelProposalBy(rng, wallettest.NewRandomAddress(rng), opts...)
}

// NewRandomChannelProposalBy creates a random channel proposal with the
// supplied options and proposer. Number of participants is fixed to 2.
func NewRandomChannelProposalBy(rng *rand.Rand, proposer wallet.Address, opts ...client.ProposalOpts) client.ChannelProposal {
	return client.NewLedgerChannelProposal(
		rng.Uint64(),
		proposer,
		channeltest.NewRandomAllocation(rng, channeltest.WithNumParts(2)),
		wallettest.NewRandomAddresses(rng, 2),
		opts...)
}
