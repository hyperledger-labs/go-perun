#!/bin/sh

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

ABIGEN=abigen
SOLC=solc

# Download solc.
if [ `uname` == "Linux" ]; then
    # GNU Linux
    wget -nc "https://github.com/ethereum/solidity/releases/download/v0.7.4/solc-static-linux"
    chmod +x solc-static-linux
    SOLC=./solc-static-linux

elif [ `uname` == "Darwin" ]; then
    # Mac OSX
    curl -L "https://github.com/ethereum/solidity/releases/download/v0.7.4/solc-macos" -o solc-macos
    chmod +x solc-macos
    SOLC=./solc-macos

else
    # Unsupported
    echo "Unsupported operating system: `uname`. Exiting."
    exit 1
fi

echo "Exec 'git submodule update --init --recursive' once after cloning."
echo "Ensure that the newest version of abigen is installed."

# Generates optimized golang bindings and runtime binaries for sol contracts.
# $1  solidity file path, relative to ../contracts/contracts/.
# $2  golang package name.
function generate() {
    FILE=$1; PKG=$2; CONTRACT=$FILE
    echo "generate package $PKG"

    mkdir -p $PKG

    # generate abi
    $ABIGEN --pkg $PKG --sol ../contracts/contracts/$FILE.sol --out $PKG/$FILE.go --solc $SOLC

    # generate go bindings
    $SOLC --bin-runtime --optimize --allow-paths *, ../contracts/contracts/$FILE.sol --overwrite -o $PKG/

    # generate binary runtime
    BIN_RUNTIME=`cat ${PKG}/${CONTRACT}.bin-runtime`
    OUT_FILE="$PKG/${CONTRACT}BinRuntime.go"
    echo "package $PKG // import \"perun.network/go-perun/backend/ethereum/bindings/$PKG\"" > $OUT_FILE
    echo >> $OUT_FILE
    echo "// ${CONTRACT}BinRuntime is the runtime part of the compiled bytecode used for deploying new contracts." >> $OUT_FILE
    echo "var ${CONTRACT}BinRuntime = \"$BIN_RUNTIME\"" >> $OUT_FILE
}

# Adjudicator
generate "Adjudicator" "adjudicator"

# Asset holders
generate "AssetHolder" "assetholder"
generate "AssetHolderETH" "assetholdereth"
generate "AssetHolderERC20" "assetholdererc20"

# Tokens
generate "PerunToken" "peruntoken"

# Applications
generate "TrivialApp" "trivialapp"

echo "Generated bindings"
