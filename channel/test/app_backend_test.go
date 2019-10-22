// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// This file contains a dummy implementation of channel.AppBackend for testing
// purposes.
// THE NAME OF THIS FILE MUST END ON `_test.go` in order to avoid calls to
// `channel.SetAppBackend()` when the package containing this file is being
// imported.

package test

import (
	"io"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

// noAppBackend implements channel.AppBackend. backend.sim.channel.AppBackend
// cannot be used because it would introduce a cyclic import dependency since
// this package uses channel.test.Asset.
type noAppBackend struct{}

var _ channel.AppBackend = &noAppBackend{}

func (noAppBackend) AppFromDefinition(addr wallet.Address) (channel.App, error) {
	return NewNoApp(addr), nil
}

func (noAppBackend) DecodeAsset(r io.Reader) (channel.Asset, error) {
	var asset Asset
	return &asset, asset.Decode(r)
}

func init() {
	channel.SetAppBackend(noAppBackend{})
}
