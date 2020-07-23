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
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/backend/ethereum/bindings/assets"
	"perun.network/go-perun/log"
)

const deployGasLimit = 6600000

// DeployETHAssetholder deploys a new ETHAssetHolder contract.
func DeployETHAssetholder(ctx context.Context, backend ContractBackend, adjudicatorAddr common.Address) (common.Address, error) {
	auth, err := backend.NewTransactor(ctx, big.NewInt(0), deployGasLimit)
	if err != nil {
		return common.Address{}, errors.WithMessage(err, "could not create transactor")
	}
	addr, tx, _, err := assets.DeployAssetHolderETH(auth, backend, adjudicatorAddr)
	if err != nil {
		return common.Address{}, errors.WithMessage(err, "could not create transaction")
	}
	if err := confirmDeployment(ctx, backend, tx); err != nil {
		return common.Address{}, errors.WithMessage(err, "deploying ethassetholder")
	}
	log.Infof("Successfully deployed AssetHolderETH at %v.", addr.Hex())
	return addr, nil
}

// DeployAdjudicator deploys a new Adjudicator contract.
func DeployAdjudicator(ctx context.Context, backend ContractBackend) (common.Address, error) {
	auth, err := backend.NewTransactor(ctx, big.NewInt(0), deployGasLimit)
	if err != nil {
		return common.Address{}, errors.WithMessage(err, "could not create transactor")
	}
	addr, tx, _, err := adjudicator.DeployAdjudicator(auth, backend)
	if err != nil {
		return common.Address{}, errors.WithMessage(err, "could not create transaction")
	}
	if err = confirmDeployment(ctx, backend, tx); err != nil {
		return common.Address{}, errors.WithMessage(err, "deploying adjudicator")
	}
	log.Infof("Successfully deployed Adjudicator at %v.", addr.Hex())
	return addr, nil
}

func confirmDeployment(ctx context.Context, backend ContractBackend, tx *types.Transaction) error {
	_, err := bind.WaitDeployed(ctx, backend, tx)
	return errors.Wrap(err, "could not execute transaction")
}
