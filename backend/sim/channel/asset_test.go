// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel // import "perun.network/go-perun/backend/sim/channel"

import (
	"fmt"
	"io"
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wire"
)

// asset simulates a `perunchannel.asset` by only containing a `string` as Name
type asset struct {
	Name string
}

var _ channel.Asset = new(asset)

// newRandomAsset returns a new random Asset
func newRandomAsset(rng *rand.Rand) *asset {
	return &asset{Name: fmt.Sprintf("Asset #%d", rng.Int63())}
}

// Encode encodes an Asset into the io.Writer `w`
func (a *asset) Encode(w io.Writer) error {
	return wire.ByteSlice(a.Name).Encode(w)
}

// Decode is not implemented in this simulation
func (a *asset) Decode(r io.Reader) error {
	log.Panic("Asset.Decode is not implemented")
	return nil
}
