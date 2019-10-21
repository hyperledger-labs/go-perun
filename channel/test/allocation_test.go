// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"io"
	"math/big"
	"testing"

	"perun.network/go-perun/channel"
	perunio "perun.network/go-perun/pkg/io"
	ioTest "perun.network/go-perun/pkg/io/test"
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

func TestAllocationSerialization(t *testing.T) {
	channel.SetAppBackend(noAppBackend{})

	inputs := []perunio.Serializable{
		&channel.Allocation{
			Assets:  []channel.Asset{&Asset{0}},
			OfParts: [][]channel.Bal{[]channel.Bal{big.NewInt(123)}},
			Locked:  []channel.SubAlloc{},
		},
		&channel.Allocation{
			Assets: []channel.Asset{&Asset{1}},
			OfParts: [][]channel.Bal{
				[]channel.Bal{big.NewInt(1)},
			},
			Locked: []channel.SubAlloc{
				channel.SubAlloc{
					ID:   channel.ID{0},
					Bals: []channel.Bal{big.NewInt(2)}},
			},
		},
		&channel.Allocation{
			Assets: []channel.Asset{&Asset{1}, &Asset{2}, &Asset{3}},
			OfParts: [][]channel.Bal{
				[]channel.Bal{big.NewInt(1), big.NewInt(10), big.NewInt(100)},
				[]channel.Bal{big.NewInt(7), big.NewInt(11), big.NewInt(13)},
			},
			Locked: []channel.SubAlloc{
				channel.SubAlloc{
					ID: channel.ID{0},
					Bals: []channel.Bal{
						big.NewInt(1), big.NewInt(3), big.NewInt(5),
					},
				},
			},
		},
	}

	ioTest.GenericSerializableTest(t, inputs...)
}
