// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	perunio "perun.network/go-perun/pkg/io"
	ioTest "perun.network/go-perun/pkg/io/test"
)

func assets(n uint) []channel.Asset {
	as := make([]channel.Asset, n)
	for i := uint(0); i < n; i++ {
		as[i] = new(test.Asset)
	}
	return as
}

func TestAllocationSerialization(t *testing.T) {
	inputs := []perunio.Serializable{
		&channel.Allocation{
			Assets:  assets(1),
			OfParts: [][]channel.Bal{[]channel.Bal{big.NewInt(123)}},
			Locked:  []channel.SubAlloc{},
		},
		&channel.Allocation{
			Assets: assets(1),
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
			Assets: assets(3),
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

func TestAllocationSerializationLimits(t *testing.T) {
	inputs := []struct {
		numAssets         int
		numParts          int
		numSuballocations int
	}{
		{channel.MaxNumAssets + 1, 1, 0},
		{1, channel.MaxNumParts + 1, 0},
		{1, 1, channel.MaxNumSuballocations + 1},
		{
			channel.MaxNumAssets + 2,
			2 * channel.MaxNumParts,
			4 * channel.MaxNumSuballocations},
	}

	for _, x := range inputs {
		allocation := &channel.Allocation{
			Assets:  make([]channel.Asset, x.numAssets),
			OfParts: make([][]channel.Bal, x.numParts),
			Locked:  make([]channel.SubAlloc, x.numSuballocations),
		}

		for i := range allocation.Assets {
			allocation.Assets[i] = &test.Asset{ID: 1}
		}

		for i := range allocation.OfParts {
			allocation.OfParts[i] = make([]channel.Bal, x.numAssets)

			for j := range allocation.OfParts[i] {
				bal := big.NewInt(int64(x.numAssets)*int64(i) + int64(j))
				allocation.OfParts[i][j] = bal
			}
		}

		for i := range allocation.Locked {
			allocation.Locked[i] = channel.SubAlloc{
				ID:   channel.ID{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)},
				Bals: make([]channel.Bal, x.numAssets)}

			for j := range allocation.Locked[i].Bals {
				bal := big.NewInt(int64(x.numAssets)*int64(i) + int64(j) + 1)
				allocation.Locked[i].Bals[j] = bal
			}
		}

		buffer := new(bytes.Buffer)
		err := allocation.Encode(buffer)
		assert.Errorf(t, err, "expected error for parameters %v", x)
	}
}
