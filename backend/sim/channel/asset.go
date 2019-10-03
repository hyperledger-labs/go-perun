// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel // import "perun.network/go-perun/backend/sim/channel"

import (
	"io"
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire"
)

// Asset simulates a `perunchannel.Asset` by only containing an `ID`
type Asset struct {
	ID int64
}

var _ channel.Asset = new(Asset)

// NewRandomAsset returns a new random Asset
func NewRandomAsset(rng *rand.Rand) *Asset {
	return &Asset{ID: rng.Int63()}
}

// Encode encodes an Asset into the io.Writer `w`
func (a *Asset) Encode(w io.Writer) error {
	return wire.Encode(w, a.ID)
}

// Decode is not implemented in this simulation
func (a *Asset) Decode(r io.Reader) error {
	return wire.Decode(r, &a.ID)
}
