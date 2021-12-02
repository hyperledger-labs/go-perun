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

package keystore_test

import (
	"bytes"
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet/test"
	pkgtest "polycry.pt/poly-go/test"
)

var dataToSign = []byte("SomeLongDataThatShouldBeSignedPlease")

const (
	invalidAddr = "123456"
)

func TestGenericSignatureTests(t *testing.T) {
	setup := newSetup(t, pkgtest.Prng(t))
	test.TestAccountWithWalletAndBackend(t, setup)
	test.GenericSignatureSizeTest(t, setup)
}

func TestGenericAddressTests(t *testing.T) {
	test.TestAddress(t, newSetup(t, pkgtest.Prng(t)))
}

func TestWallet_Contains(t *testing.T) {
	rng := pkgtest.Prng(t)
	w := ethwallettest.NewTmpWallet()

	assert.False(t, w.Contains(ethwallet.AsEthAddr(test.NewRandomAddress(rng))),
		"Expected wallet not to contain an empty account")

	acc := w.NewAccount()
	assert.True(t, w.Contains(ethwallet.AsEthAddr(acc.Address())), "Expected wallet to contain account")
}

func TestSignatures(t *testing.T) {
	acc := ethwallettest.NewTmpWallet().NewAccount()
	sign, err := acc.SignData(dataToSign)
	assert.NoError(t, err, "Sign with new account should succeed")
	assert.Equal(t, len(sign), ethwallet.SigLen, "Ethereum signature has wrong length")
	valid, err := new(ethwallet.Backend).VerifySignature(dataToSign, sign, acc.Address())
	assert.True(t, valid, "Verification should succeed")
	assert.NoError(t, err, "Verification should succeed")
}

func TestBackend(t *testing.T) {
	backend := new(ethwallet.Backend)

	s := newSetup(t, pkgtest.Prng(t))

	buff := bytes.NewReader(s.AddressEncoded)
	addr := backend.NewAddress()
	err := io.Decode(buff, addr)
	assert.NoError(t, err, "NewAddress from Bytes should work")

	buff = bytes.NewReader([]byte(invalidAddr))
	addr = backend.NewAddress()
	err = io.Decode(buff, addr)
	assert.Error(t, err, "Conversion from wrong address should fail")
}

func newSetup(t require.TestingT, prng *rand.Rand) *test.Setup {
	w := ethwallettest.NewTmpWallet()

	addressNotInWallet := ethwallettest.NewRandomAddress(prng)
	var buff bytes.Buffer
	err := io.Encode(&buff, &addressNotInWallet)
	if err != nil {
		panic(err)
	}
	addrEncoded := buff.Bytes()

	require.NoError(t, err, "decoding valid address should not fail")

	return &test.Setup{
		Wallet:          w,
		AddressInWallet: w.NewAccount().Address(),
		Backend:         new(ethwallet.Backend),
		AddressEncoded:  addrEncoded,
		ZeroAddress:     ethwallet.AsWalletAddr(common.Address{}),
		DataToSign:      dataToSign,
	}
}

func TestCurve_SigningAndVerifying(t *testing.T) {
	msg, err := hex.DecodeString("f27b90711d11d10a155fc8ba0eed1ffbf449cf3730d88c0cb77b98f61750ab34000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000022000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000160000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000014000000000000000000000000000000000000000000000000000000000000000010000000000000000000000002c2b9c9a4a25e24b174f26114e8926a9f2128fe40000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000de0b6b3a7640000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err, "decode msg should not error")
	sig, err := hex.DecodeString("538da6430f7915832de165f89c69239020461b80861559a00d4f5a2a7705765219eb3969eb7095f8addb6bf9c9f96f6adf44cfd4a8136516f88b337a428bf1bb1b")
	require.NoError(t, err, "decode sig should not error")
	addr := ethwallet.Address(common.HexToAddress("f17f52151EbEF6C7334FAD080c5704D77216b732"))
	b, err := ethwallet.VerifySignature(msg, sig, &addr)
	assert.NoError(t, err, "VerifySignature should not error")
	assert.True(t, b, "VerifySignature")
}
