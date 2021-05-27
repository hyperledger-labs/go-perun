// Copyright 2021 - See NOTICE file for copyright holders.
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

package bindings

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/backend/ethereum/bindings/assetholder"
	"perun.network/go-perun/backend/ethereum/bindings/assetholdererc20"
	"perun.network/go-perun/backend/ethereum/bindings/assetholdereth"
	"perun.network/go-perun/backend/ethereum/bindings/peruntoken"
	"perun.network/go-perun/backend/ethereum/bindings/trivialapp"
)

// This file contains all the parsed ABI definitions of our contracts.
// Use it together with `bind.NewBoundContract` to create a bound contract.

var (
	// ERC20TokenABI is the parsed ABI definition of contract ERC20Token.
	ERC20TokenABI abi.ABI
	// AdjudicatorABI is the parsed ABI definition of contract Adjudicator.
	AdjudicatorABI abi.ABI
	// AssetHolderABI is the parsed ABI definition of contract AssetHolder.
	AssetHolderABI abi.ABI
	// ETHAssetHolderABI is the parsed ABI definition of contract ETHAssetHolder.
	ETHAssetHolderABI abi.ABI
	// ERC20AssetHolderABI is the parsed ABI definition of contract ERC20AssetHolder.
	ERC20AssetHolderABI abi.ABI
	// TrivialAppABI is the parsed ABI definition of contract TrivialApp.
	TrivialAppABI abi.ABI
)

func init() {
	parseABI := func(raw string) abi.ABI {
		abi, err := abi.JSON(strings.NewReader(raw))
		if err != nil {
			panic(err)
		}
		return abi
	}

	ERC20TokenABI = parseABI(peruntoken.ERC20ABI)
	AdjudicatorABI = parseABI(adjudicator.AdjudicatorABI)
	AssetHolderABI = parseABI(assetholder.AssetHolderABI)
	ETHAssetHolderABI = parseABI(assetholdereth.AssetHolderETHABI)
	ERC20AssetHolderABI = parseABI(assetholdererc20.AssetHolderERC20ABI)
	TrivialAppABI = parseABI(trivialapp.TrivialAppABI)
}
