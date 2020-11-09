package assets

import (
	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// runtimePlaceholder indicates constructor variables in runtime binary code.
const runtimePlaceholder = "7f0000000000000000000000000000000000000000000000000000000000000000"

func AssetHolderERC20BinRuntimeFor(token common.Address) string {
	tokenHex := hex.EncodeToString(token[:])
	return strings.ReplaceAll(AssetHolderERC20BinRuntime,
		runtimePlaceholder,
		runtimePlaceholder[:len(runtimePlaceholder)-len(tokenHex)]+tokenHex)
}
