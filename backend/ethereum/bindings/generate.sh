#!/bin/bash

# Copyright 2020 - See NOTICE file for copyright holders.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

# Download solc.
wget -nc "https://github.com/ethereum/solidity/releases/download/v0.7.3/solc-static-linux"
chmod +x solc-static-linux
echo -e "Exec 'git submodule update --init --recursive' once after cloning."
echo -e "Ensure that the newest version of abigen is installed"

# Generates optimized golang bindings and runtime binaries for sol contracts.
# $1  solidity file path, relative to ../contracts/contracts/.
# $1  golang package name.
# $2… list of contract names.
function generate() {
    file=$1; pkg=$2
    shift; shift   # skip the first two args.
    for contract in "$@"; do
        abigen --pkg $pkg --sol ../contracts/contracts/$file.sol --out $pkg/$file.go --solc ./solc-static-linux
        ./solc-static-linux --bin-runtime --optimize --allow-paths *, ../contracts/contracts/$file.sol --overwrite -o $pkg/
        echo -e "package $pkg\n\n // ${contract}BinRuntime is the runtime part of the compiled bytecode used for deploying new contracts.\nvar ${contract}BinRuntime = \`$(<${pkg}/${contract}.bin-runtime)\`" > "$pkg/${contract}BinRuntime.go"
    done
}

# Adjudicator
generate "Adjudicator" "adjudicator" "Adjudicator"

# PerunToken, AssetHolderETH and AssetHolderERC20
# Pragma statements can only be in a solidity file once, so
# we remove the duplicates with awk.
cat ../contracts/contracts/PerunToken.sol \
    <(awk '/^pragma/{p=1;next}{if(p){print}}' ../contracts/contracts/AssetHolderETH.sol) \
    <(awk '/^pragma/{p=1;next}{if(p){print}}' ../contracts/contracts/AssetHolderERC20.sol) \
    > ../contracts/contracts/Contracts.sol
generate "Contracts" "assets" "PerunToken" "AssetHolderETH" "AssetHolderERC20"
rm ../contracts/contracts/Contracts.sol

abigen --version --solc ./solc-static-linux
echo -e "Generated bindings"
