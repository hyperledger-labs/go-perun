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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/assets"
)

// ETHDepositor deposits funds into the `AssetHolderETH` contract.
// It has no state and is therefore completely reusable.
type ETHDepositor struct{}

// ETHDepositorGasLimit is the limit of Gas that an `ETHDepositor` will spend
// when depositing funds. It is set to 50000.
// A `Deposit` call uses ~47kGas on average.
const ETHDepositorGasLimit = 50000

// Deposit returns the transactions for despositing Ethereum or an error otherwise.
func (d *ETHDepositor) Deposit(ctx context.Context, req DepositReq) (types.Transactions, error) {
	// Bind an `AssetHolderETH` instance. Using `AssetHolder` is also possible
	// since we only use the interface functions here.
	contract, err := assets.NewAssetHolderETH(common.Address(req.Asset), req.CB)
	if err != nil {
		return nil, errors.WithMessagef(err, "binding AssetHolderETH contract at: %x", req.Asset)
	}
	opts, err := req.CB.NewTransactor(ctx, ETHDepositorGasLimit, req.Account)
	if err != nil {
		return nil, errors.WithMessagef(err, "creating transactor for asset: %x", req.Asset)
	}
	opts.Value = req.Balance

	tx, err := contract.Deposit(opts, req.FundingID, req.Balance)
	return []*types.Transaction{tx}, errors.WithMessage(err, "AssetHolderETH depositing")
}

// NumTX returns 1 since it only does Deposit.
func (*ETHDepositor) NumTX() uint32 {
	return 1
}
