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
	Tr         channel.Transactor
	ValidAcc   accounts.Account // wallet should contain key corresponding to this account.
	MissingAcc accounts.Account // wallet should not contain key corresponding to this account.
}

// GenericTransactorTest provides generic transactor tests.
func GenericTransactorTest(t *testing.T, setup TransactorSetup) {
	t.Run("happy", func(t *testing.T) {
		transactOpts, err := setup.Tr.NewTransactor(setup.ValidAcc)
		require.NoError(t, err)

		data := []byte("some random tx data")
		rawTx := types.NewTransaction(uint64(1), common.Address{}, big.NewInt(1), uint64(1), big.NewInt(1), data)
		signer := types.HomesteadSigner{}
		signedTx, err := transactOpts.Signer(signer, setup.ValidAcc.Address, rawTx)
		assert.NoError(t, err)
		assert.NotNil(t, signedTx)

		txHash := signer.Hash(rawTx).Bytes()
		v, r, s := signedTx.RawSignatureValues()
		sig := append(r.Bytes(), s.Bytes()...)
		sig = append(sig, v.Bytes()...)
		sig[64] -= 27

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

		data := []byte("some random tx data")
		rawTx := types.NewTransaction(uint64(1), common.Address{}, big.NewInt(1), uint64(1), big.NewInt(1), data)
		signer := types.HomesteadSigner{}
		_, err = transactOpts.Signer(signer, setup.MissingAcc.Address, rawTx)
		assert.Error(t, err)
	})
}
