// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"io"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

// NoAppBackend implements channel.AppBackend. backend.sim.channel.AppBackend
// cannot be used because it would introduce a cyclic import dependency since
// this package uses channel.test.Asset.
type NoAppBackend struct{}

var _ channel.AppBackend = &NoAppBackend{}

func (NoAppBackend) AppFromDefinition(addr wallet.Address) (channel.App, error) {
	return NewNoApp(addr), nil
}

func (NoAppBackend) DecodeAsset(r io.Reader) (channel.Asset, error) {
	var asset Asset
	return &asset, asset.Decode(r)
}
