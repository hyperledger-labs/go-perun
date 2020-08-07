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

package bindings

//go:generate wget -nc "https://github.com/ethereum/solidity/releases/download/v0.7.0/solc-static-linux"
//go:generate chmod +x solc-static-linux
//go:generate echo -e "Ensure that the newest version of abigen is installed"
//go:generate echo -e "Make sure you cloned the contracts submodule (git submodule update --init)"
//go:generate abigen --pkg adjudicator --sol ../contracts/contracts/Adjudicator.sol --out adjudicator/Adjudicator.go --solc ./solc-static-linux
//go:generate ./solc-static-linux --bin-runtime --optimize --allow-paths *, ../contracts/contracts/Adjudicator.sol --overwrite -o adjudicator/
//go:generate bash -c "echo -e \"package adjudicator\n\n// AdjudicatorBinRuntime is the runtime part of the compiled bytecode used for deploying new contracts.\nvar AdjudicatorBinRuntime = \\\"$(<adjudicator/Adjudicator.bin-runtime)\\\"\" > adjudicator/AdjudicatorBinRuntime.go"
//go:generate abigen --pkg assets --sol ../contracts/contracts/AssetHolderETH.sol --out assets/AssetHolderETH.go --solc ./solc-static-linux
//go:generate ./solc-static-linux --bin-runtime --optimize --allow-paths *, ../contracts/contracts/AssetHolderETH.sol --overwrite -o assets/
//go:generate bash -c "echo -e \"package assets\n\n// AssetHolderETHBinRuntime is the runtime part of the compiled bytecode used for deploying new contracts.\nvar AssetHolderETHBinRuntime = \\\"$(<assets/AssetHolderETH.bin-runtime)\\\"\" > assets/AssetHolderETHBinRuntime.go"
//go:generate abigen --version --solc ./solc-static-linux
//go:generate echo -e "Generated bindings"
