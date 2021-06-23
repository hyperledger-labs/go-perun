// Copyright 2019 - See NOTICE file for copyright holders.
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
	"context"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/backend/ethereum/channel"
	ethchanneltest "perun.network/go-perun/backend/ethereum/channel/test"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/channel/test"
	iotest "perun.network/go-perun/pkg/io/test"
	pkgtest "perun.network/go-perun/pkg/test"
	perunwallet "perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
)

func TestAdjudicator_PureFunctions(t *testing.T) {
	rng := pkgtest.Prng(t)
	s := ethchanneltest.NewSimSetup(rng)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	adj, err := channel.DeployAdjudicator(ctx, *s.CB, s.TxSender.Account)
	require.NoError(t, err)
	contr, err := adjudicator.NewAdjudicator(adj, *s.CB)
	require.NoError(t, err)
	opts := &bind.CallOpts{Context: ctx}

	t.Run("testCalcID", func(t *testing.T) {
		testCalcID(t, rng, contr, opts)
	})
	t.Run("testHashState", func(t *testing.T) {
		testHashState(t, rng, contr, opts)
	})
}

func testCalcID(t *testing.T, rng *rand.Rand, contr *adjudicator.Adjudicator, opts *bind.CallOpts) {
	for i := 0; i < 100; i++ {
		params := test.NewRandomParams(rng)
		ethParams := channel.ToEthParams(params)
		ethId, err := contr.ChannelID(opts, ethParams)
		require.NoError(t, err)
		chID := channel.CalcID(params)

		require.NoError(t, err)
		require.Equal(t, chID, ethId)
	}

	assert.Panics(t, func() {
		channel.CalcID(nil)
	})
}

func testHashState(t *testing.T, rng *rand.Rand, contr *adjudicator.Adjudicator, opts *bind.CallOpts) {
	for i := 0; i < 100; i++ {
		state := test.NewRandomState(rng)
		ethState := channel.ToEthState(state)
		ethHash, err := contr.HashState(opts, ethState)
		require.NoError(t, err)
		stateHash := channel.HashState(state)

		require.NoError(t, err)
		require.Equal(t, stateHash, ethHash)
	}

	assert.Panics(t, func() {
		channel.HashState(nil)
	})
}

func TestGenericTests(t *testing.T) {
	setup := newChannelSetup(pkgtest.Prng(t))
	test.GenericBackendTest(t, setup)
	test.GenericStateEqualTest(t, setup.State, setup.State2)
}

func newChannelSetup(rng *rand.Rand) *test.Setup {
	params, state := test.NewRandomParamsAndState(rng, test.WithNumLocked(int(rng.Int31n(4)+1)))
	params2, state2 := test.NewRandomParamsAndState(rng, test.WithIsFinal(!state.IsFinal), test.WithNumLocked(int(rng.Int31n(4)+1)))

	createAddr := func() perunwallet.Address {
		return wallettest.NewRandomAddress(rng)
	}

	return &test.Setup{
		Params:        params,
		Params2:       params2,
		State:         state,
		State2:        state2,
		Account:       wallettest.NewRandomAccount(rng),
		RandomAddress: createAddr,
	}
}

func TestAssetSerialization(t *testing.T) {
	rng := pkgtest.Prng(t)
	asset := ethwallettest.NewRandomAddress(rng)
	reader, writer := io.Pipe()
	done := make(chan struct{})

	go func() {
		defer close(done)
		assert.NoError(t, asset.Encode(writer))
	}()

	backend := new(channel.Backend)
	asset2, err := backend.DecodeAsset(reader)
	assert.NoError(t, err, "Decode asset should not produce error")
	assert.Equal(t, &asset, asset2, "Decode asset should return the initial asset")
	<-done

	for i := 0; i < 10; i++ {
		asset := ethwallettest.NewRandomAddress(rng)
		iotest.GenericSerializerTest(t, &asset)
	}
}
