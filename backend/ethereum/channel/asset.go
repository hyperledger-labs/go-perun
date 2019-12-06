// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"io"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire"
)

// Asset is an Ethereum asset
type Asset struct {
	common.Address
}

var _ channel.Asset = new(Asset)

// NewRandomAsset returns a new random sim Asset
func NewRandomAsset(rng *rand.Rand) *Asset {
	addr := wallet.NewRandomAddress(rng)
	return &Asset{addr.Address}
}

// Encode encodes a sim Asset into the io.Writer `w`
func (a Asset) Encode(w io.Writer) error {
	return wire.Encode(w, a.Bytes())
}

// Decode decodes a sim Asset from the io.Reader `r`
func (a *Asset) Decode(r io.Reader) error {
	return wire.Decode(r, &a)
}
