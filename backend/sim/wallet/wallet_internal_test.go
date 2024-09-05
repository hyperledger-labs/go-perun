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

package wallet

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/wallet/test"
	pkgtest "polycry.pt/poly-go/test"
)

// TestSignatureSerialize tests serializeSignature and deserializeSignature since
// a signature is only a []byte, we cant use io.serializer here.
func TestSignatureSerialize(t *testing.T) {
	a := assert.New(t)
	// Constant seed for determinism
	rng := pkgtest.Prng(t)

	// More iterations are better for catching value dependent bugs
	for i := 0; i < 10; i++ {
		rBytes := make([]byte, 32)
		sBytes := make([]byte, 32)

		// These always return nil error
		rng.Read(rBytes)
		rng.Read(sBytes)

		r := new(big.Int).SetBytes(rBytes)
		s := new(big.Int).SetBytes(sBytes)

		sig, err1 := serializeSignature(r, s)
		a.Nil(err1, "Serialization should not fail")
		a.Equal(curve.Params().BitSize/4, len(sig), "Signature has wrong size")
		R, S, err2 := deserializeSignature(sig)

		a.Nil(err2, "Deserialization should not fail")
		a.Equal(r, R, "Serialized and deserialized r values should be equal")
		a.Equal(s, S, "Serialized and deserialized s values should be equal")
	}
}

func TestGenericTests(t *testing.T) {
	t.Run("Generic Address Test", func(t *testing.T) {
		t.Parallel()
		rng := pkgtest.Prng(t, "address")
		test.TestAddress(t, newWalletSetup(rng))
	})
	t.Run("Generic Signature Test", func(t *testing.T) {
		t.Parallel()
		rng := pkgtest.Prng(t, "signature")
		test.TestAccountWithWalletAndBackend(t, newWalletSetup(rng))
		test.GenericSignatureSizeTest(t, newWalletSetup(rng))
	})

	// NewRandomAddress is also tested in channel_test but since they are two packages,
	// we also need to test it here
	rng := pkgtest.Prng(t)
	for i := 0; i < 10; i++ {
		addr0 := NewRandomAddress(rng)
		addr1 := NewRandomAddress(rng)
		assert.NotEqual(
			t, addr0, addr1, "Two random accounts should not be the same")

		addrStrLen := addrLen*2 + 2 // hex encoded and prefixed with 0x
		str0 := addr0.String()
		str1 := addr1.String()
		assert.Equal(
			t, addrStrLen, len(str0), "First address '%v' has wrong length", str0)
		assert.Equal(
			t, addrStrLen, len(str1), "Second address '%v' has wrong length", str1)
		assert.NotEqual(
			t, str0, str1, "Printed addresses are unlikely to be identical")
	}
}

func newWalletSetup(rng *rand.Rand) *test.Setup {
	w := NewWallet()

	data := make([]byte, 128)
	_, err := rng.Read(data)
	if err != nil {
		panic(err)
	}

	addressNotInWallet := NewRandomAccount(rng).Address()
	addrMarshalled, err := addressNotInWallet.MarshalBinary()
	if err != nil {
		panic(err)
	}

	zeroAddr := &Address{
		Curve: curve,
		X:     big.NewInt(0),
		Y:     big.NewInt(0),
	}

	return &test.Setup{
		Backend:           new(Backend),
		Wallet:            w,
		AddressInWallet:   w.NewRandomAccount(rng).Address(),
		AddressMarshalled: addrMarshalled,
		ZeroAddress:       zeroAddr,
		DataToSign:        data,
	}
}
