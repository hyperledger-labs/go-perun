package adjudicator

//go:generate echo -e "\\e[01;31mEnsure that solc version is ^0.5.13\\e[0m"
//go:generate echo -e "\\e[01;31mMake sure you cloned the contracts submodule (git submodule update --init)\\e[0m"
//go:generate solc --bin-runtime --optimize ../contracts/contracts/Adjudicator.sol --overwrite -o ./
//go:generate solc --bin-runtime --optimize ../contracts/contracts/AssetHolderETH.sol --overwrite -o ./
//go:generate abigen --pkg adjudicator --sol ../contracts/contracts/Adjudicator.sol --out adjudicator/Adjudicator.go
//go:generate abigen --pkg assets --sol ../contracts/contracts/AssetHolderETH.sol --out assets/AssetHolderETH.go
//go:generate abigen --version
//go:generate echo -e "\\e[01;31mGenerated bindings\\e[0m"
