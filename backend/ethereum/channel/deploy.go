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

package channel

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/backend/ethereum/bindings/assetholdererc20"
	"perun.network/go-perun/backend/ethereum/bindings/assetholdereth"
	"perun.network/go-perun/backend/ethereum/bindings/peruntoken"
	"perun.network/go-perun/backend/ethereum/bindings/trivialapp"
	"perun.network/go-perun/client"
	"perun.network/go-perun/log"
	pcontext "perun.network/go-perun/pkg/context"
)

const deployGasLimit = 6600000

// DeployPerunToken deploys a new PerunToken contract.
// Returns txTimedOutError if the context is cancelled or if the context
// deadline is exceeded when waiting for the transaction to be mined.
func DeployPerunToken(ctx context.Context, backend ContractBackend, deployer accounts.Account, initAccs []common.Address, initBals *big.Int) (common.Address, error) {
	return deployContract(ctx, backend, deployer, "PerunToken",
		func(auth *bind.TransactOpts, cb ContractBackend) (common.Address, *types.Transaction, error) {
			addr, tx, _, err := peruntoken.DeployPerunToken(auth, backend, initAccs, initBals)
			return addr, tx, err
		})
}

// DeployETHAssetholder deploys a new ETHAssetHolder contract.
// Returns txTimedOutError if the context is cancelled or if the context
// deadline is exceeded when waiting for the transaction to be mined.
func DeployETHAssetholder(ctx context.Context, backend ContractBackend, adjudicatorAddr common.Address, deployer accounts.Account) (common.Address, error) {
	return deployContract(ctx, backend, deployer, "ETHAssetHolder",
		func(auth *bind.TransactOpts, cb ContractBackend) (common.Address, *types.Transaction, error) {
			addr, tx, _, err := assetholdereth.DeployAssetHolderETH(auth, cb, adjudicatorAddr)
			return addr, tx, err
		})
}

// DeployERC20Assetholder deploys a new ERC20AssetHolder contract.
// Returns txTimedOutError if the context is cancelled or if the context
// deadline is exceeded when waiting for the transaction to be mined.
func DeployERC20Assetholder(ctx context.Context, backend ContractBackend, adjudicatorAddr common.Address, tokenAddr common.Address, deployer accounts.Account) (common.Address, error) {
	return deployContract(ctx, backend, deployer, "ERC20AssetHolder",
		func(auth *bind.TransactOpts, cb ContractBackend) (common.Address, *types.Transaction, error) {
			addr, tx, _, err := assetholdererc20.DeployAssetHolderERC20(auth, backend, adjudicatorAddr, tokenAddr)
			return addr, tx, err
		})
}

// DeployAdjudicator deploys a new Adjudicator contract.
// Returns txTimedOutError if the context is cancelled or if the context
// deadline is exceeded when waiting for the transaction to be mined.
func DeployAdjudicator(ctx context.Context, backend ContractBackend, deployer accounts.Account) (common.Address, error) {
	return deployContract(ctx, backend, deployer, "Adjudicator",
		func(auth *bind.TransactOpts, cb ContractBackend) (common.Address, *types.Transaction, error) {
			addr, tx, _, err := adjudicator.DeployAdjudicator(auth, backend)
			return addr, tx, err
		})
}

// DeployTrivialApp deploys a new TrivialApp contract.
// Returns txTimedOutError if the context is cancelled or if the context
// deadline is exceeded when waiting for the transaction to be mined.
func DeployTrivialApp(ctx context.Context, backend ContractBackend, deployer accounts.Account) (common.Address, error) {
	return deployContract(ctx, backend, deployer, "TrivialApp",
		func(auth *bind.TransactOpts, cb ContractBackend) (common.Address, *types.Transaction, error) {
			addr, tx, _, err := trivialapp.DeployTrivialApp(auth, backend)
			return addr, tx, err
		})
}

// Returns txTimedOutError if the context is cancelled or if the context
// deadline is exceeded when waiting for the transaction to be mined.
func deployContract(ctx context.Context, cb ContractBackend, deployer accounts.Account, name string, f func(*bind.TransactOpts, ContractBackend) (common.Address, *types.Transaction, error)) (common.Address, error) {
	auth, err := cb.NewTransactor(ctx, deployGasLimit, deployer)
	if err != nil {
		return common.Address{}, errors.WithMessage(err, "creating transactor")
	}
	addr, tx, err := f(auth, cb)
	if err != nil {
		err = checkIsChainNotReachableError(err)
		return common.Address{}, errors.WithMessage(err, "creating transaction")
	}
	if _, err := bind.WaitDeployed(ctx, cb, tx); err != nil {
		switch {
		case pcontext.IsContextError(err):
			txType := fmt.Sprintf("deploy %s", name)
			err = client.NewTxTimedoutError(txType, tx.Hash().Hex(), err.Error())
		case isChainNotReachableError(err):
			err = client.NewChainNotReachableError(err)
		default:
			err = errors.WithStack(err)
		}
		return common.Address{}, errors.WithMessagef(err, "deploying %s", name)
	}
	log.Infof("Deployed %s at %v.", name, addr.Hex())
	return addr, nil
}
