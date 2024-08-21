// Copyright 2022 - See NOTICE file for copyright holders.
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

package multi

import (
	"fmt"

	"perun.network/go-perun/channel"
)

type (
	// Asset defines a multi-ledger asset.
	Asset interface {
		channel.Asset
		AssetID() AssetID
	}

	AssetID struct {
		BackendID uint32
		LedgerId  LedgerID
	}

	// LedgerIDMapKey is the map key representation of a ledger identifier.
	LedgerIDMapKey string

	// LedgerID represents a ledger identifier.
	LedgerID interface {
		MapKey() LedgerIDMapKey
	}

	assets []channel.Asset
)

// LedgerIDs returns the identifiers of the associated ledgers.
func (a assets) LedgerIDs() ([]AssetID, error) {
	var result []AssetID
	for _, asset := range a {
		ma, ok := asset.(Asset)
		if !ok {
			return nil, fmt.Errorf("wrong asset type: expected Asset, got %T", asset)
		}

		assetID := ma.AssetID()

		// Append the pair of IDs to the result slice
		result = append(result, assetID)
	}

	return result, nil
}

// IsMultiLedgerAssets returns whether the assets are from multiple ledgers.
func IsMultiLedgerAssets(assets []channel.Asset) bool {
	hasMulti := false
	var id AssetID
	for _, asset := range assets {
		multiAsset, ok := asset.(Asset)
		switch {
		case !ok:
			continue
		case !hasMulti:
			hasMulti = true
			id = multiAsset.AssetID()
		case id.LedgerId.MapKey() != multiAsset.AssetID().LedgerId.MapKey():
			return true
		}
	}
	return false
}
