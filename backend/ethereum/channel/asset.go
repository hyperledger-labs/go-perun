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
	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/assetholder"
	"perun.network/go-perun/backend/ethereum/bindings/assetholdererc20"
	"perun.network/go-perun/backend/ethereum/bindings/assetholdereth"
	cherrors "perun.network/go-perun/backend/ethereum/channel/errors"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
)

// Asset is an Ethereum asset.
type Asset = wallet.Address

var _ channel.Asset = new(Asset)

// ValidateAssetHolderETH checks if the bytecode at the given asset holder ETH
// address is correct and if the adjudicator address is correctly set in the
// asset holder contract. The contract code at the adjudicator address is not
// validated, it is the user's responsibility to provide a valid adjudicator
// address.
//
// Returns a ContractBytecodeError if the bytecode is invalid. This error can
// be checked with function IsErrInvalidContractCode.
func ValidateAssetHolderETH(ctx context.Context,
	backend bind.ContractBackend, assetHolderETH, adjudicator common.Address) error {
	return validateAssetHolder(ctx, backend, assetHolderETH, adjudicator,
		assetholdereth.AssetHolderETHBinRuntime)
}

// ValidateAssetHolderERC20 checks if the bytecode at the given asset holder
// ERC20 address is correct and if the adjudicator address is correctly set in
// the asset holder contract. The contract code at the adjudicator address is
// not validated, it is the user's responsibility to provide a valid
// adjudicator address.
//
// Returns a ContractBytecodeError if the bytecode is invalid. This error can
// be checked with function IsErrInvalidContractCode.
func ValidateAssetHolderERC20(ctx context.Context,
	backend bind.ContractBackend, assetHolderERC20, adjudicator, token common.Address) error {
	return validateAssetHolder(ctx, backend, assetHolderERC20, adjudicator,
		assetHolderERC20BinRuntimeFor(token))
}

func validateAssetHolder(ctx context.Context,
	backend bind.ContractBackend, assetHolderAddr, adjudicatorAddr common.Address, bytecode string) error {
	if err := validateContract(ctx, backend, assetHolderAddr, bytecode); err != nil {
		return errors.WithMessage(err, "validating asset holder")
	}

	assetHolder, err := assetholder.NewAssetHolder(assetHolderAddr, backend)
	if err != nil {
		return errors.Wrap(err, "binding AssetHolder")
	}
	opts := bind.CallOpts{
		Pending: false,
		Context: ctx,
	}
	if addrSetInContract, err := assetHolder.Adjudicator(&opts); err != nil {
		err = cherrors.CheckIsChainNotReachableError(err)
		return errors.WithMessage(err, "fetching adjudicator address set in asset holder contract")
	} else if addrSetInContract != adjudicatorAddr {
		return errors.Wrap(ErrInvalidContractCode, "incorrect adjudicator code")
	}

	return nil
}

func validateContract(ctx context.Context,
	backend bind.ContractCaller, contract common.Address, bytecode string) error {
	code, err := backend.CodeAt(ctx, contract, nil)
	if err != nil {
		err = cherrors.CheckIsChainNotReachableError(err)
		return errors.WithMessage(err, "fetching contract code")
	}
	if hex.EncodeToString(code) != bytecode {
		return errors.Wrap(ErrInvalidContractCode, "incorrect contract code")
	}
	return nil
}

func assetHolderERC20BinRuntimeFor(token common.Address) string {
	// runtimePlaceholder indicates constructor variables in runtime binary code.
	const runtimePlaceholder = "7f0000000000000000000000000000000000000000000000000000000000000000"

	tokenHex := hex.EncodeToString(token[:])
	return strings.ReplaceAll(assetholdererc20.AssetHolderERC20BinRuntime,
		runtimePlaceholder,
		runtimePlaceholder[:len(runtimePlaceholder)-len(tokenHex)]+tokenHex)
}
