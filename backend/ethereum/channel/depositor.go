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

package channel

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/core/types"

	"perun.network/go-perun/channel"
)

type (
	// Depositor is used by the `Funder` to deposit funds for ledger channel
	// funding. Depositor are reusable such that one Depositor per asset is enough.
	Depositor interface {
		// Deposit returns the transactions needed for the deposit or an error.
		// The transactions should already be sent to the chain, such that
		// `abi/bind.WaitMined` can be used to await their success.
		// When one of the TX fails, the status of the following ones is ignored.
		Deposit(context.Context, DepositReq) ([]*types.Transaction, error)

		// NumTX returns how many transactions a `Deposit` call needs.
		NumTX() uint32
	}

	// DepositReq contains all necessary data for a `Depositor` to deposit funds.
	// It is much smaller than a `FundingReq` and only holds the information
	// for one Funding-ID.
	DepositReq struct {
		Balance   channel.Bal      // How much should be deposited.
		CB        ContractBackend  // Used to bind contracts and send TX.
		Account   accounts.Account // Depositor's account.
		Asset     Asset            // Address of the AssetHolder.
		FundingID [32]byte         // Needed by the AssetHolder.
	}
)

// NewDepositReq returns a new `DepositReq`.
func NewDepositReq(balance channel.Bal, cb ContractBackend, asset Asset, account accounts.Account, fundingID [32]byte) *DepositReq {
	return &DepositReq{
		Balance:   balance,
		CB:        cb,
		Account:   account,
		Asset:     asset,
		FundingID: fundingID,
	}
}
