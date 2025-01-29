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

package payment

import (
	"math/rand"

	"perun.network/go-perun/wallet"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
)

// Randomizer implements channel.test.AppRandomizer.
type Randomizer struct{}

var _ test.AppRandomizer = (*Randomizer)(nil)

// NewRandomApp always returns a payment app with a different address.
func (*Randomizer) NewRandomApp(rng *rand.Rand, bID wallet.BackendID) channel.App {
	return &App{test.NewRandomAppID(rng, bID)}
}

// NewRandomData returns NoData because a PaymentApp does not have data.
func (*Randomizer) NewRandomData(*rand.Rand) channel.Data {
	return channel.NoData()
}
