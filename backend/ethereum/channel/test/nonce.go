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
