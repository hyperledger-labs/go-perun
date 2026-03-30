// Copyright 2026 - See NOTICE file for copyright holders.
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
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	_ "perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/channel"
	ctest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestTransactionDecDecodeUsesParticipantBackends(t *testing.T) {
	rng := pkgtest.Prng(t)
	accs, addrs := wallettest.NewRandomAccounts(rng, 2, channel.TestBackendID)
	params := ctest.NewRandomParams(rng, ctest.WithParts(addrs))
	state := ctest.NewRandomState(rng, ctest.WithID(params.ID()), ctest.WithNumParts(len(addrs)))

	sigs := make([]wallet.Sig, len(addrs))
	for i := range addrs {
		sig, err := channel.Sign(accs[i][channel.TestBackendID], state, channel.TestBackendID)
		require.NoError(t, err)
		sigs[i] = sig
	}

	var encoded bytes.Buffer
	original := channel.Transaction{State: state, Sigs: sigs}
	require.NoError(t, original.Encode(&encoded))

	var decoded channel.Transaction
	require.NoError(t, (channel.TransactionDec{Tx: &decoded, Parts: addrs}).Decode(&encoded))
	require.Equal(t, original, decoded)
}
