// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package channel

import (
	"io"
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire"
)

// Asset simulates a `channel.Asset` by only containing an `ID`
type Asset struct {
	ID int64
}

var _ channel.Asset = new(Asset)

// NewRandomAsset returns a new random sim Asset
func NewRandomAsset(rng *rand.Rand) *Asset {
	return &Asset{ID: rng.Int63()}
}

// Encode encodes a sim Asset into the io.Writer `w`
func (a Asset) Encode(w io.Writer) error {
	return wire.Encode(w, a.ID)
}

// Decode decodes a sim Asset from the io.Reader `r`
func (a *Asset) Decode(r io.Reader) error {
	return wire.Decode(r, &a.ID)
}
