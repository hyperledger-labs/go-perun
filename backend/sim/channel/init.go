// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel // import "perun.network/go-perun/backend/sim/channel"

import (
	"io"

	"perun.network/go-perun/channel"
	test "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
)

func init() {
	channel.SetBackend(new(backend))
	channel.SetAppBackend(new(noAppBackend))
}

type noAppBackend struct{}

var _ channel.AppBackend = &noAppBackend{}

func (noAppBackend) AppFromDefinition(addr wallet.Address) (channel.App, error) {
	return test.NewNoApp(addr), nil
}

func (noAppBackend) DecodeAsset(r io.Reader) (channel.Asset, error) {
	var asset test.Asset
	return &asset, asset.Decode(r)
}
