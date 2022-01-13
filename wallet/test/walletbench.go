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

package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// GenericAccountBenchmark runs a suite designed to benchmark the general speed of an implementation of an Account.
// This function should be called by every implementation of the Account interface.
func GenericAccountBenchmark(b *testing.B, s *Setup) {
	b.Helper()
	b.Run("Sign", func(b *testing.B) { benchAccountSign(b, s) })
}

func benchAccountSign(b *testing.B, s *Setup) {
	b.Helper()
	perunAcc, err := s.Wallet.Unlock(s.AddressInWallet)
	require.Nil(b, err)

	for n := 0; n < b.N; n++ {
		_, err := perunAcc.SignData(s.DataToSign)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// GenericBackendBenchmark runs a suite designed to benchmark the general speed
// of an implementation of a Backend.
//
// This function should be called by every implementation of the Backend
// interface.
func GenericBackendBenchmark(b *testing.B, s *Setup) {
	b.Helper()
	b.Run("VerifySig", func(b *testing.B) { benchBackendVerifySig(b, s) })
	b.Run("UnmarshalAddress", func(b *testing.B) { benchUnmarshalAddress(b, s) })
}

func benchBackendVerifySig(b *testing.B, s *Setup) {
	b.Helper()
	// We dont want to measure the SignDataWithPW here, just need it for the verification
	b.StopTimer()
	perunAcc, err := s.Wallet.Unlock(s.AddressInWallet)
	require.Nil(b, err)
	signature, err := perunAcc.SignData(s.DataToSign)
	require.Nil(b, err)
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		ok, err := s.Backend.VerifySignature(s.DataToSign, signature, perunAcc.Address())

		if !ok {
			b.Fatal(err)
		}
	}
}

func benchUnmarshalAddress(b *testing.B, s *Setup) {
	b.Helper()
	for n := 0; n < b.N; n++ {
		addr := s.Backend.NewAddress()
		err := addr.UnmarshalBinary(s.AddressMarshalled)
		if err != nil {
			b.Fatal(err)
		}
	}
}
