// Copyright 2025 - See NOTICE file for copyright holders.
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
	"math/rand"

	"perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/channel"
)

// AppID represents an app identifier.
type AppID struct {
	*wallet.Address
}

// Equal returns whether the object is equal to the given object.
func (id AppID) Equal(b channel.AppID) bool {
	bTyped, ok := b.(AppID)
	if !ok {
		return false
	}

	return id.Address.Equal(bTyped.Address)
}

// Key returns the key representation of this app identifier.
func (id AppID) Key() channel.AppIDKey {
	b, err := id.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return channel.AppIDKey(b)
}

// NewRandomAppID generates a new random app identifier.
func NewRandomAppID(rng *rand.Rand) AppID {
	addr := wallet.NewRandomAddress(rng)
	return AppID{Address: addr}
}
