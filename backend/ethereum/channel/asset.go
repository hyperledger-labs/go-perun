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

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/assets"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
)

// Asset is an Ethereum asset.
type Asset = wallet.Address

var _ channel.Asset = new(Asset)

// ValidateAssetHolderETH checks if the bytecode at given addresses are correct,
// and if the adjudicator address is correctly set in the asset holder contract.
// Returns a ContractBytecodeError if the bytecode at given address is invalid.
// This error can be checked with function IsErrInvalidContractCode.
func ValidateAssetHolderETH(ctx context.Context,
	backend bind.ContractBackend, assetHolderETH, adjudicatorAddr common.Address) error {
	code, err := backend.CodeAt(ctx, assetHolderETH, nil)
	if err != nil {
		return errors.Wrap(err, "fetching AssetHolderETH")
	}
	if hex.EncodeToString(code) != assets.AssetHolderETHBinRuntime {
		return errors.Wrap(ErrInvalidContractCode, "incorrect AssetHolderETH code")
	}

	assetHolder, err := assets.NewAssetHolderETH(assetHolderETH, backend)
	if err != nil {
		return errors.Wrap(err, "binding AssetHolderETH")
	}
	opts := bind.CallOpts{
		Pending: false,
		Context: ctx,
	}
	if addrSetInContract, err := assetHolder.Adjudicator(&opts); err != nil {
		return errors.Wrap(err, "fetching adjudicator address set in asset holder contract")
	} else if addrSetInContract != adjudicatorAddr {
		return errors.Wrap(ErrInvalidContractCode, "incorrect adjudicator code")
	}

	return ValidateAdjudicator(ctx, backend, adjudicatorAddr)
}
