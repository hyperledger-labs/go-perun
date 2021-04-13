#!/usr/bin/env sh

# Copyright 2021 - See NOTICE file for copyright holders.
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

# Define ABIGEN and SOLC default values.
ABIGEN="${ABIGEN-abigen}"
SOLC="${SOLC-solc}"

if ! $ABIGEN --version
then
    echo "Please make abigen v1.9.25+ available through PATH or set the environment variable AGIBEN."
    exit 1
fi

if ! $SOLC --version
then
    echo "Please make abigen v0.7.4 available through PATH or set the environment variable SOLC."
    exit 1
fi

echo "Please ensure that the repository was cloned with submodules: 'git submodule update --init --recursive'."

# Generates optimized golang bindings and runtime binaries for sol contracts.
# $1  solidity file path, relative to ../contracts/contracts/.
# $2  golang package name.
generate() {
    NAME=$1; PKG=$2
    echo "Generating $PKG bindings..."

    rm -r $PKG
    mkdir $PKG

    # Compile
    $SOLC --abi --bin --bin-runtime --optimize --allow-paths ../contracts/vendor, ../contracts/contracts/$NAME.sol -o $PKG/

    # Generate bindings
    $ABIGEN --abi $PKG/$NAME.abi --bin $PKG/$NAME.bin --pkg $PKG --out $PKG/$NAME.go

    # Generate binary runtime
    BIN_RUNTIME=`cat ${PKG}/${NAME}.bin-runtime`
    OUT_FILE="$PKG/${NAME}BinRuntime.go"
    echo "package $PKG // import \"perun.network/go-perun/backend/ethereum/bindings/$PKG\"" > $OUT_FILE
    echo >> $OUT_FILE
    echo "// ${NAME}BinRuntime is the runtime part of the compiled bytecode used for deploying new contracts." >> $OUT_FILE
    echo "var ${NAME}BinRuntime = \"$BIN_RUNTIME\"" >> $OUT_FILE
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

echo "Bindings generated successfully."
