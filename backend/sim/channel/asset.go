// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package channel

import (
	"io"
	"math/rand"

	"perun.network/go-perun/channel"
	perunio "perun.network/go-perun/pkg/io"
)

// Asset simulates a `channel.Asset` by only containing an `ID`.
type Asset struct {
	ID int64
}

var _ channel.Asset = new(Asset)

// NewRandomAsset returns a new random sim Asset.
func NewRandomAsset(rng *rand.Rand) *Asset {
	return &Asset{ID: rng.Int63()}
}

// Encode encodes a sim Asset into the io.Writer `w`.
func (a Asset) Encode(w io.Writer) error {
	return perunio.Encode(w, a.ID)
}

// Decode decodes a sim Asset from the io.Reader `r`.
func (a *Asset) Decode(r io.Reader) error {
	return perunio.Decode(r, &a.ID)
}
