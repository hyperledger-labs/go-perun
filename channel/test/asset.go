// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/channel/test"

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
func newRandomAsset(rng *rand.Rand) *Asset {
	return &Asset{ID: rng.Int63()}
}

// Encode encodes an Asset into the io.Writer `w`
func (a *Asset) Encode(w io.Writer) error {
	return wire.Encode(w, a.ID)
}

// Decode decodes an Asset from the io.Reader `r`
func (a *Asset) Decode(r io.Reader) error {
	return wire.Decode(r, &a.ID)
}
