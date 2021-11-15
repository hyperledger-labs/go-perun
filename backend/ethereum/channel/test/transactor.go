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
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/channel"
)

// TransactorSetup holds the setup for running generic tests on a transactor implementation.
type TransactorSetup struct {
	Signer     types.Signer
	ChainID    int64
	Tr         channel.Transactor
	ValidAcc   accounts.Account // wallet should contain key corresponding to this account.
	MissingAcc accounts.Account // wallet should not contain key corresponding to this account.
}

// GenericSignerTest tests that a transactor produces the correct signatures
// for the passed signer.
func GenericSignerTest(t *testing.T, rng *rand.Rand, setup TransactorSetup) {
	t.Helper()
	signer := setup.Signer
	chainID := setup.ChainID
	data := make([]byte, rng.Int31n(100)+1)
	rng.Read(data)

	t.Run("happy", func(t *testing.T) {
		transactOpts, err := setup.Tr.NewTransactor(setup.ValidAcc)
		require.NoError(t, err)
		rawTx := types.NewTransaction(uint64(1), common.Address{}, big.NewInt(1), uint64(1), big.NewInt(1), data)
		signedTx, err := transactOpts.Signer(setup.ValidAcc.Address, rawTx)
		assert.NoError(t, err)
		require.NotNil(t, signedTx)

		txHash := signer.Hash(rawTx).Bytes()
		v, r, s := signedTx.RawSignatureValues()
		sig := sigFromRSV(t, r, s, v, chainID)
		pk, err := crypto.SigToPub(txHash, sig)
		require.NoError(t, err)
		addr := crypto.PubkeyToAddress(*pk)
		assert.Equal(t, setup.ValidAcc.Address.Bytes(), addr.Bytes())
	})

	t.Run("missing_account", func(t *testing.T) {
		_, err := setup.Tr.NewTransactor(setup.MissingAcc)
		assert.Error(t, err)
	})

	t.Run("wrong_sender", func(t *testing.T) {
		transactOpts, err := setup.Tr.NewTransactor(setup.ValidAcc)
		require.NoError(t, err)

		rawTx := types.NewTransaction(uint64(1), common.Address{}, big.NewInt(1), uint64(1), big.NewInt(1), data)
		_, err = transactOpts.Signer(setup.MissingAcc.Address, rawTx)
		assert.Error(t, err)
	})
}

func sigFromRSV(t *testing.T, r, s, _v *big.Int, chainID int64) []byte {
	t.Helper()
	sig := make([]byte, 65)
	copy(sig[32-len(r.Bytes()):32], r.Bytes())
	copy(sig[64-len(s.Bytes()):64], s.Bytes())
	v := byte(_v.Uint64()) // Needed for chain ids > 110.

	if chainID == 0 {
		sig[64] = v - 27
	} else {
		sig[64] = v - byte(chainID*2+35) // Underflow is ok here.
	}
	require.Contains(t, []byte{0, 1}, sig[64], "Invalid v")
	return sig
}
