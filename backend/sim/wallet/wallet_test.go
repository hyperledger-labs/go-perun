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

package wallet_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/sim/wallet"
	pkgtest "polycry.pt/poly-go/test"
)

func TestWallet_AddAccount(t *testing.T) {
	rng := pkgtest.Prng(t)
	w := wallet.NewWallet()
	acc := wallet.NewRandomAccount(rng)

	assert.False(t, w.HasAccount(acc))
	assert.NoError(t, w.AddAccount(acc))
	assert.True(t, w.HasAccount(acc))
	assert.Error(t, w.AddAccount(acc))
}

func TestWallet_Unlock(t *testing.T) {
	rng := pkgtest.Prng(t)

	// Create a wallet from existing accounts, as if just restored. These
	// accounts are initially locked.
	acc := wallet.NewRandomAccount(rng)
	w := wallet.NewRestoredWallet(acc)

	t.Run("sign before unlock", func(t *testing.T) {
		sig, err := acc.SignData([]byte("----"))
		require.Error(t, err)
		require.Nil(t, sig)
	})

	t.Run("unlock", func(t *testing.T) {
		testAcc, err := w.Unlock(acc.Address())
		require.NoError(t, err)
		require.Same(t, acc, testAcc)
	})

	t.Run("sign after unlock", func(t *testing.T) {
		sig, err := acc.SignData([]byte("----"))
		require.NoError(t, err)
		require.NotNil(t, sig)
	})

	t.Run("redundant unlock", func(t *testing.T) {
		testAcc, err := w.Unlock(acc.Address())
		require.NoError(t, err)
		require.Same(t, acc, testAcc)
	})

	w.LockAll()
	t.Run("after LockAll", func(t *testing.T) {
		sig, err := acc.SignData([]byte("----"))
		require.Error(t, err)
		require.Nil(t, sig)
	})

	t.Run("unknown unlock", func(t *testing.T) {
		acc, err := w.Unlock(wallet.NewRandomAddress(rng))
		assert.Error(t, err)
		assert.Nil(t, acc)
	})
}

func TestWallet_UsageCounting(t *testing.T) {
	rng := pkgtest.Prng(t)

	w := wallet.NewWallet()
	const N = 10

	acc := w.NewRandomAccount(rng).(*wallet.Account)
	assert.Zero(t, w.UsageCount(acc.Address()))

	t.Run("unmatched decrement", func(t *testing.T) {
		acc := w.NewRandomAccount(rng).(*wallet.Account)
		assert.Panics(t, func() { w.DecrementUsage(acc.Address()) })
		assert.Equal(t, w.UsageCount(acc.Address()), -1)
		assert.True(t, w.HasAccount(acc))
	})

	t.Run("increment", func(t *testing.T) {
		for i := 1; i <= N; i++ {
			assert.NotPanics(t, func() { w.IncrementUsage(acc.Address()) })
			assert.Equal(t, w.UsageCount(acc.Address()), i)
			assert.True(t, w.HasAccount(acc))
		}
	})

	t.Run("decrement", func(t *testing.T) {
		for i := N - 1; i >= 0; i-- {
			assert.True(t, w.HasAccount(acc))
			assert.NotPanics(t, func() { w.DecrementUsage(acc.Address()) })
			if i > 0 {
				assert.Equal(t, w.UsageCount(acc.Address()), i)
			} else {
				assert.Panics(t, func() { w.UsageCount(acc.Address()) })
				assert.False(t, w.HasAccount(acc))
			}
		}
	})

	t.Run("removed", func(t *testing.T) {
		assert.False(t, w.HasAccount(acc))
	})

	t.Run("invalid address", func(t *testing.T) {
		assert.Panics(t, func() { w.IncrementUsage(wallet.NewRandomAddress(rng)) })
		assert.Panics(t, func() { w.DecrementUsage(wallet.NewRandomAddress(rng)) })
	})
}
