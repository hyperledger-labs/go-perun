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

package test

import (
	"math/rand"

	"perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
)

// NewRandomAddress returns a new random peer address. Currently still a stub
// until the crypto for peer addresses is decided.
func NewRandomAddress(rng *rand.Rand) wire.Address {
	return test.NewRandomAddress(rng)
}

// NewRandomAddresses returns a slice of random peer addresses.
func NewRandomAddresses(rng *rand.Rand, n int) []wire.Address {
	walletAddresses := test.NewRandomAddresses(rng, n)
	addresses := make([]wire.Address, len(walletAddresses))
	for i, x := range walletAddresses {
		addresses[i] = x
	}
	return addresses
}

// NewRandomEnvelope returns an envelope around message m with random sender and
// recipient generated using randomness from rng.
func NewRandomEnvelope(rng *rand.Rand, m wire.Msg) *wire.Envelope {
	return &wire.Envelope{
		Sender:    test.NewRandomAddress(rng),
		Recipient: test.NewRandomAddress(rng),
		Msg:       m,
	}
}
