// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test // import "perun.network/go-perun/wallet/test"

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

// GenericAccountBenchmark runs a suite designed to benchmark the general speed of an implementation of an Account.
// This function should be called by every implementation of the Account interface.
func GenericAccountBenchmark(b *testing.B, s *Setup) {
	b.Run("Sign", func(b *testing.B) { benchAccountSign(b, s) })
}

func benchAccountSign(b *testing.B, s *Setup) {
	perunAcc, err := s.UnlockedAccount()
	require.Nil(b, err)

	for n := 0; n < b.N; n++ {
		_, err := perunAcc.SignData(s.DataToSign)

		if err != nil {
			b.Fatal(err)
		}
	}
}

// GenericWalletBenchmark runs a suite designed to benchmark the general speed of an implementation of a Wallet.
// This function should be called by every implementation of the Wallet interface.
func GenericWalletBenchmark(b *testing.B, s *Setup) {
	b.Run("Conn&Disconn", func(b *testing.B) { benchWalletConnectAndDisconnect(b, s) })
	b.Run("Connect", func(b *testing.B) { benchWalletConnect(b, s) })
	b.Run("Accounts", func(b *testing.B) { benchWalletAccounts(b, s) })
	b.Run("Contains", func(b *testing.B) { benchWalletContains(b, s) })
}

func benchWalletConnect(b *testing.B, s *Setup) {
	for n := 0; n < b.N; n++ {
		err := s.InitWallet(s.Wallet)

		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchWalletConnectAndDisconnect(b *testing.B, s *Setup) {
	for n := 0; n < b.N; n++ {
		err := s.InitWallet(s.Wallet)

		if err != nil {
			b.Fatal(err)
		}

		err = s.Wallet.Disconnect()

		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchWalletContains(b *testing.B, s *Setup) {
	account, err := s.UnlockedAccount()
	require.Nil(b, err)

	for n := 0; n < b.N; n++ {
		in := s.Wallet.Contains(account)

		if !in {
			b.Fatal("address not found")
		}
	}
}

func benchWalletAccounts(b *testing.B, s *Setup) {
	require.Nil(b, s.InitWallet(s.Wallet))

	for n := 0; n < b.N; n++ {
		accounts := s.Wallet.Accounts()

		if len(accounts) != 1 {
			b.Fatal("there was not exactly one account in the wallet")
		}
	}
}

// GenericBackendBenchmark runs a suite designed to benchmark the general speed of an implementation of a Backend.
// This function should be called by every implementation of the Backend interface.
func GenericBackendBenchmark(b *testing.B, s *Setup) {
	b.Run("VerifySig", func(b *testing.B) { benchBackendVerifySig(b, s) })
	b.Run("DecodeAddress", func(b *testing.B) { benchBackendDecodeAddress(b, s) })
}

func benchBackendVerifySig(b *testing.B, s *Setup) {
	// We dont want to measure the SignDataWithPW here, just need it for the verification
	b.StopTimer()
	perunAcc, err := s.UnlockedAccount()
	require.Nil(b, err)
	signature, err := perunAcc.SignData(s.DataToSign)
	require.Nil(b, err)
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		ok, err := s.Backend.VerifySignature(s.DataToSign, signature, perunAcc.Address())

		if ok != true {
			b.Fatal(err)
		}
	}
}

func benchBackendDecodeAddress(b *testing.B, s *Setup) {
	for n := 0; n < b.N; n++ {
		_, err := s.Backend.DecodeAddress(bytes.NewReader(s.AddressBytes))

		if err != nil {
			b.Fatal(err)
		}
	}
}
