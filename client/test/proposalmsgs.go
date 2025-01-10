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

package test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
	pkgtest "polycry.pt/poly-go/test"
)

// ProposalMsgsSerializationTest runs serialization tests on proposal messages.
func ProposalMsgsSerializationTest(t *testing.T, serializerTest func(t *testing.T, msg wire.Msg)) {
	t.Helper()
	channelProposalReqSerializationTest(t, serializerTest)
	channelProposalAccSerializationTest(t, serializerTest)
	channelProposalRejSerializationTest(t, serializerTest)
}

func channelProposalReqSerializationTest(t *testing.T, serializerTest func(t *testing.T, msg wire.Msg)) {
	t.Helper()
	rng := pkgtest.Prng(t)
	for i := 0; i < 16; i++ {
		var (
			app client.ProposalOpts
			m   wire.Msg
			err error
		)
		if i&1 == 0 {
			app = client.WithApp(test.NewRandomAppAndData(rng))
		}
		switch i % 3 {
		case 0:
			m = NewRandomLedgerChannelProposal(rng, client.WithNonceFrom(rng), app)
		case 1:
			m, err = NewRandomSubChannelProposal(rng, client.WithNonceFrom(rng), app)
			require.NoError(t, err)
		case 2: //nolint: gomnd 	// This is not a magic number.
			m, err = NewRandomVirtualChannelProposal(rng, client.WithNonceFrom(rng), app)
			require.NoError(t, err)
		}
		serializerTest(t, m)
	}
}

func channelProposalAccSerializationTest(t *testing.T, serializerTest func(t *testing.T, msg wire.Msg)) {
	t.Helper()
	rng := pkgtest.Prng(t)
	t.Run("ledger channel", func(t *testing.T) {
		for i := 0; i < 16; i++ {
			proposal := NewRandomLedgerChannelProposal(rng)
			m := proposal.Accept(wallettest.NewRandomAddresses(rng, 0), client.WithNonceFrom(rng))
			serializerTest(t, m)
		}
	})
	t.Run("sub channel", func(t *testing.T) {
		for i := 0; i < 16; i++ {
			var err error
			proposal, err := NewRandomSubChannelProposal(rng)
			require.NoError(t, err)
			m := proposal.Accept(client.WithNonceFrom(rng))
			serializerTest(t, m)
		}
	})
	t.Run("virtual channel", func(t *testing.T) {
		for i := 0; i < 16; i++ {
			var err error
			proposal, err := NewRandomVirtualChannelProposal(rng)
			require.NoError(t, err)
			m := proposal.Accept(wallettest.NewRandomAddresses(rng, 0))
			serializerTest(t, m)
		}
	})
}

func channelProposalRejSerializationTest(t *testing.T, serializerTest func(t *testing.T, msg wire.Msg)) {
	t.Helper()
	minLen := 16
	maxLenDiff := 16
	rng := pkgtest.Prng(t)
	for i := 0; i < 16; i++ {
		m := &client.ChannelProposalRejMsg{
			ProposalID: newRandomProposalID(rng),
			Reason:     newRandomASCIIString(rng, minLen, maxLenDiff),
		}
		serializerTest(t, m)
	}
}

func newRandomProposalID(rng *rand.Rand) (id client.ProposalID) {
	rng.Read(id[:])
	return
}

// newRandomASCIIString returns a random ascii string of length between minLen and
// minLen+maxLenDiff.
func newRandomASCIIString(rng *rand.Rand, minLen, maxLenDiff int) string {
	str := make([]byte, minLen+rng.Intn(maxLenDiff))
	const firstPrintableASCII = 32
	const lastPrintableASCII = 126
	for i := range str {
		str[i] = byte(firstPrintableASCII + rng.Intn(lastPrintableASCII-firstPrintableASCII))
	}
	return string(str)
}
