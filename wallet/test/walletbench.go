// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/wallet/test"

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// GenericAccountBenchmark runs a suite designed to benchmark the general speed of an implementation of an Account.
// This function should be called by every implementation of the Account interface.
func GenericAccountBenchmark(b *testing.B, s *Setup) {
	require.Nil(b, s.InitWallet(s.Wallet))
	b.Run("Sign", func(t *testing.B) { benchAccountSign(t, s) })
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
	b.Run("Conn&Disconn", func(t *testing.B) { benchWalletConnectAndDisconnect(t, s) })
	b.Run("Connect", func(t *testing.B) { benchWalletConnect(t, s) })
	b.Run("Accounts", func(t *testing.B) { benchWalletAccounts(t, s) })
	b.Run("Contains", func(t *testing.B) { benchWalletContains(t, s) })
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
	b.Run("VerifySig", func(t *testing.B) { benchBackendVerifySig(t, s) })
	b.Run("FromString", func(t *testing.B) { benchBackendNewAddressFromString(t, s) })
	b.Run("FromBytes", func(t *testing.B) { benchBackendNewAddressFromBytes(t, s) })
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

func benchBackendNewAddressFromString(b *testing.B, s *Setup) {
	for n := 0; n < b.N; n++ {
		_, err := s.Backend.NewAddressFromString(s.AddrString)

		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchBackendNewAddressFromBytes(b *testing.B, s *Setup) {
	data, err := s.Backend.NewAddressFromString(s.AddrString)
	bytes := data.Bytes()

	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		_, err := s.Backend.NewAddressFromBytes(bytes)

		if err != nil {
			b.Fatal(err)
		}
	}
}
