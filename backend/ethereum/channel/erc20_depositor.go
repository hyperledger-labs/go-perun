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

	"perun.network/go-perun/backend/ethereum/bindings/assetholdererc20"
	"perun.network/go-perun/backend/ethereum/bindings/peruntoken"
)

// ERC20Depositor deposits tokens into the `AssetHolderERC20` contract.
// It is bound to a token but can be reused to deposit multiple times.
type ERC20Depositor struct {
	Token common.Address
}

// ERC20DepositorTXGasLimit is the limit of Gas that an `ERC20Depositor` will
// spend per transaction when depositing funds.
// An `IncreaseAllowance` uses ~45kGas and a `Deposit` call ~84kGas on average.
const ERC20DepositorTXGasLimit = 100000

// Deposit deposits ERC20 tokens into the ERC20 AssetHolder specified at the
// requests's asset address.
func (d *ERC20Depositor) Deposit(ctx context.Context, req DepositReq) (types.Transactions, error) {
	// Bind a `AssetHolderERC20` instance.
	assetholder, err := assetholdererc20.NewAssetHolderERC20(common.Address(req.Asset), req.CB)
	if err != nil {
		return nil, errors.WithMessagef(err, "binding AssetHolderERC20 contract at: %x", req.Asset)
	}
	// Bind an `ERC20` instance.
	token, err := peruntoken.NewERC20(d.Token, req.CB)
	if err != nil {
		return nil, errors.WithMessagef(err, "binding ERC20 contract at: %x", d.Token)
	}
	// Increase the allowance.
	opts, err := req.CB.NewTransactor(ctx, ERC20DepositorTXGasLimit, req.Account)
	if err != nil {
		return nil, errors.WithMessagef(err, "creating transactor for asset: %x", req.Asset)
	}
	tx1, err := token.IncreaseAllowance(opts, common.Address(req.Asset), req.Balance)
	if err != nil {
		return nil, errors.WithMessagef(err, "increasing allowance for asset: %x", req.Asset)
	}
	// Deposit.
	opts, err = req.CB.NewTransactor(ctx, ERC20DepositorTXGasLimit, req.Account)
	if err != nil {
		return nil, errors.WithMessagef(err, "creating transactor for asset: %x", req.Asset)
	}
	tx2, err := assetholder.Deposit(opts, req.FundingID, req.Balance)
	return []*types.Transaction{tx1, tx2}, errors.WithMessage(err, "AssetHolderERC20 depositing")
}

// NumTX returns 2 since it does IncreaseAllowance and Deposit.
func (*ERC20Depositor) NumTX() uint32 {
	return 2
}
