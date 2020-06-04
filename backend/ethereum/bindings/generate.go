// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package bindings

//go:generate wget -nc "https://github.com/ethereum/solidity/releases/download/v0.5.16/solc-static-linux"
//go:generate chmod +x solc-static-linux
//go:generate echo -e "Ensure that the newest version of abigen is installed"
//go:generate echo -e "Make sure you cloned the contracts submodule (git submodule update --init)"
//go:generate abigen --pkg adjudicator --sol ../contracts/contracts/Adjudicator.sol --out adjudicator/Adjudicator.go --solc ./solc-static-linux
//go:generate ./solc-static-linux --bin-runtime --optimize ../contracts/contracts/Adjudicator.sol --overwrite -o adjudicator/
//go:generate bash -c "echo -e \"package adjudicator\n\n// AdjudicatorBinRuntime is the runtime part of the compiled bytecode used for deploying new contracts.\nvar AdjudicatorBinRuntime = \\\"$(<adjudicator/Adjudicator.bin-runtime)\\\"\" > adjudicator/AdjudicatorBinRuntime.go"
//go:generate abigen --pkg assets --sol ../contracts/contracts/AssetHolderETH.sol --out assets/AssetHolderETH.go --solc ./solc-static-linux
//go:generate ./solc-static-linux --bin-runtime --optimize ../contracts/contracts/AssetHolderETH.sol --overwrite -o assets/
//go:generate bash -c "echo -e \"package assets\n\n// AssetHolderETHBinRuntime is the runtime part of the compiled bytecode used for deploying new contracts.\nvar AssetHolderETHBinRuntime = \\\"$(<assets/AssetHolderETH.bin-runtime)\\\"\" > assets/AssetHolderETHBinRuntime.go"
//go:generate abigen --version --solc ./solc-static-linux
//go:generate echo -e "Generated bindings"
