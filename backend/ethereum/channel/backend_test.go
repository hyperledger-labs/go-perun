// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel

import (
	"encoding/hex"
	"io"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/backend/ethereum/wallet"
	_ "perun.network/go-perun/backend/ethereum/wallet"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	iotest "perun.network/go-perun/pkg/io/test"
	perunwallet "perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
)

func TestGenericTests(t *testing.T) {
	setup := newChannelSetup()
	test.GenericBackendTest(t, setup)
}

func newChannelSetup() *test.Setup {
	rng := rand.New(rand.NewSource(1337))

	app := wallettest.NewRandomAddress(rng)
	app2 := wallettest.NewRandomAddress(rng)

	params := test.NewRandomParams(rng, app)
	params2 := test.NewRandomParams(rng, app2)

	state := test.NewRandomState(rng, params)
	state2 := test.NewRandomState(rng, params2)
	state2.IsFinal = !state.IsFinal

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

func newAddressFromString(s string) *wallet.Address {
	return &wallet.Address{Address: common.HexToAddress(s)}
}

func TestChannelID(t *testing.T) {
	tests := []struct {
		name        string
		aliceAddr   string
		bobAddr     string
		appAddr     string
		challengDur uint64
		nonceStr    string
		channelID   string
	}{
		{"Test case 1",
			"0xf17f52151EbEF6C7334FAD080c5704D77216b732",
			"0xC5fdf4076b8F3A5357c5E395ab970B5B54098Fef",
			"0x9FBDa871d559710256a2502A2517b794B482Db40",
			uint64(60),
			"B0B0FACE",
			"f27b90711d11d10a155fc8ba0eed1ffbf449cf3730d88c0cb77b98f61750ab34"},
		{"Test case 2",
			"0x0000000000000000000000000000000000000000",
			"0x0000000000000000000000000000000000000000",
			"0x0000000000000000000000000000000000000000",
			uint64(0),
			"0",
			"c8ac0e8f7eeea864a050a8626dfa0ffb916f43c90bc6b2ba68df6ed063c952e2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nonce, ok := new(big.Int).SetString(tt.nonceStr, 16)
			assert.True(t, ok, "Setting the nonce should not fail")
			alice := newAddressFromString(tt.aliceAddr)
			bob := newAddressFromString(tt.bobAddr)
			app := newAddressFromString(tt.appAddr)
			params := channel.Params{
				ChallengeDuration: tt.challengDur,
				Nonce:             nonce,
				Parts:             []perunwallet.Address{alice, bob},
				App:               channel.NewMockApp(app),
			}
			cID := channel.ChannelID(&params)
			preCalc, err := hex.DecodeString(tt.channelID)
			assert.NoError(t, err, "Decoding the channelID should not error")
			assert.Equal(t, preCalc, cID[:], "ChannelID should match the testcase")
		})
	}
}

func Test_transformPartBals(t *testing.T) {
	tests := []struct {
		name string
		args [][]*big.Int
		want [][]*big.Int
	}{
		{"Test1",
			[][]*big.Int{
				{big.NewInt(1), big.NewInt(4)},
				{big.NewInt(2), big.NewInt(3)},
				{big.NewInt(6), big.NewInt(5)},
				{big.NewInt(7), big.NewInt(9)}},
			[][]*big.Int{
				{big.NewInt(1), big.NewInt(2), big.NewInt(6), big.NewInt(7)},
				{big.NewInt(4), big.NewInt(3), big.NewInt(5), big.NewInt(9)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, transformPartBals(tt.args), tt.name)
		})
	}
}

func TestAssetSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	var asset Asset = ethwallettest.NewRandomAddress(rng)
	reader, writer := io.Pipe()
	done := make(chan struct{})

	go func() {
		defer close(done)
		assert.NoError(t, asset.Encode(writer))
	}()

	asset2, err := DecodeAsset(reader)
	assert.NoError(t, err, "Decode asset should not produce error")
	assert.Equal(t, &asset, asset2, "Decode asset should return the initial asset")
	<-done

	iotest.GenericSerializerTest(t, &asset)
}
