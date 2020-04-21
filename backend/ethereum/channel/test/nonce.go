// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"perun.network/go-perun/backend/ethereum/wallet"

	perun "perun.network/go-perun/wallet"
)

// NonceDiff returns the difference between the nonce of `address` before and after calling `f` iff no other error was encountered.
func NonceDiff(address perun.Address, ct bind.ContractTransactor, f func() error) (int, error) {
	// Get the current nonce
	oldNonce, err := ct.PendingNonceAt(context.Background(), common.Address(*address.(*wallet.Address)))
	if err != nil {
		return -1, err
	}
	// Execute the function
	fErr := f()
	// Get the new nonce
	newNonce, err := ct.PendingNonceAt(context.Background(), common.Address(*address.(*wallet.Address)))
	if err != nil {
		return -1, err
	}
	return int(newNonce - oldNonce), fErr
}
