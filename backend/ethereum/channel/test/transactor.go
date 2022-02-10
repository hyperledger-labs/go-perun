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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/wallet"
)

// TxType is a transaction type, specifying how it is hashed for signing and how
// the v value of the signature is coded.
type TxType int

const (
	// LegacyTx - legacy transaction with v = {0,1} + 27.
	LegacyTx TxType = iota
	// EIP155Tx - EIP155 transaction with v = {0,1} + CHAIN_ID * 2 + 35.
	EIP155Tx
	// EIP1559Tx - EIP1559 transaction with v = {0,1}.
	EIP1559Tx
)

// TransactorSetup holds the setup for running generic tests on a transactor implementation.
type TransactorSetup struct {
	Signer     types.Signer
	ChainID    int64
	TxType     TxType // Transaction type to generate and check against this signer
	Tr         channel.Transactor
	ValidAcc   accounts.Account // wallet should contain key corresponding to this account.
	MissingAcc accounts.Account // wallet should not contain key corresponding to this account.
}

const signerTestDataMaxLength = 100

// GenericSignerTest tests that a transactor produces the correct signatures
// for the passed signer.
func GenericSignerTest(t *testing.T, rng *rand.Rand, setup TransactorSetup) {
	t.Helper()

	newTx := func() *types.Transaction {
		data := make([]byte, rng.Int31n(signerTestDataMaxLength)+1)
		rng.Read(data)
		switch setup.TxType {
		case LegacyTx, EIP155Tx:
			return types.NewTx(&types.LegacyTx{
				Value: big.NewInt(rng.Int63()),
				Data:  data,
			})
		case EIP1559Tx:
			return types.NewTx(&types.DynamicFeeTx{
				ChainID: big.NewInt(setup.ChainID),
				Value:   big.NewInt(rng.Int63()),
				Data:    data,
			})
		}
		panic("unsupported tx type")
	}

	t.Run("happy", func(t *testing.T) {
		transactOpts, err := setup.Tr.NewTransactor(setup.ValidAcc)
		require.NoError(t, err)
		tx := newTx()
		signedTx, err := transactOpts.Signer(setup.ValidAcc.Address, tx)
		assert.NoError(t, err)
		require.NotNil(t, signedTx)

		txHash := setup.Signer.Hash(tx).Bytes()
		v, r, s := signedTx.RawSignatureValues()
		sig := sigFromRSV(t, r, s, v, &setup)
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

		_, err = transactOpts.Signer(setup.MissingAcc.Address, newTx())
		assert.Error(t, err)
	})
}

func sigFromRSV(t *testing.T, r, s, _v *big.Int, setup *TransactorSetup) []byte {
	t.Helper()
	const (
		elemLen         = 32
		sigVEIP155Shift = 35
		sigVLegacyShift = 27
	)
	var (
		sig = make([]byte, wallet.SigLen)
		rb  = r.Bytes()
		sb  = s.Bytes()
		v   = byte(_v.Uint64()) // truncation anticipated for large ChainIDs
	)
	copy(sig[elemLen-len(rb):elemLen], rb)
	copy(sig[elemLen*2-len(sb):elemLen*2], sb)

	switch setup.TxType {
	case LegacyTx:
		v -= sigVLegacyShift
	case EIP155Tx:
		v -= byte(setup.ChainID*2 + sigVEIP155Shift) // underflow anticipated
	case EIP1559Tx:
		// EIP1559 transactions simply code the y-parity into v, so no correction
		// necessary.
	}
	require.Containsf(t, []byte{0, 1}, v,
		"Invalid v (txType: %v; chainID: %d)", setup.TxType, setup.ChainID)

	sig[wallet.SigLen-1] = v
	return sig
}
