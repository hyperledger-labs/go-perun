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

	"perun.network/go-perun/wallet"

	"github.com/google/uuid"
	"perun.network/go-perun/channel"
)

// MockAppRandomizer implements the AppRandomizer interface.
type MockAppRandomizer struct {
	id uuid.UUID // Unique identifier for each instance
}

// NewMockAppRandomizer creates a new instance of MockAppRandomizer with a unique identifier.
func NewMockAppRandomizer() *MockAppRandomizer {
	return &MockAppRandomizer{id: uuid.New()}
}

// NewRandomApp creates a new MockApp with a random address.
func (MockAppRandomizer) NewRandomApp(rng *rand.Rand, bID wallet.BackendID) channel.App {
	return channel.NewMockApp(NewRandomAppID(rng, bID))
}

// NewRandomData creates a new MockOp with a random operation.
func (MockAppRandomizer) NewRandomData(rng *rand.Rand) channel.Data {
	return channel.NewMockOp(channel.MockOp(rng.Uint64()))
}

// Equal returns false for any comparison, ensuring two MockAppRandomizers are always considered different.
func (m *MockAppRandomizer) Equal(other *MockAppRandomizer) bool {
	return m.id != other.id
}
