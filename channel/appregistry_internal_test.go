// Copyright 2020 - See NOTICE file for copyright holders.
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
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/wallet"
	wtest "perun.network/go-perun/wallet/test"
	"polycry.pt/poly-go/test"
)

func TestAppRegistry(t *testing.T) {
	test.OnlyOnce(t)
	rng := test.Prng(t)

	backup := struct {
		resolvers  []appRegEntry
		singles    map[wallet.AddrKey]App
		defaultRes AppResolver
	}{
		resolvers:  appRegistry.resolvers,
		singles:    appRegistry.singles,
		defaultRes: appRegistry.defaultRes,
	}

	t.Cleanup(func() {
		appRegistry.resolvers = backup.resolvers
		appRegistry.singles = backup.singles
		appRegistry.defaultRes = backup.defaultRes
	})

	t.Run("TestAppregistryPanicsAndErrors", func(t *testing.T) {
		testAppRegistryPanicsAndErrors(t)
	})
	t.Run("TestAppRegistryIdentity", func(t *testing.T) {
		testAppRegistryIdentity(t, rng)
	})
}

func testAppRegistryPanicsAndErrors(t *testing.T) {
	resetAppRegistry()
	assert.Panics(t, func() { RegisterAppResolver(nil, nil) })
	assert.Panics(t, func() { RegisterAppResolver(func(wallet.Address) bool { return true }, nil) })
	assert.Panics(t, func() { RegisterAppResolver(nil, &MockAppResolver{}) })

	assert.Panics(t, func() { RegisterApp(nil) })
	assert.Panics(t, func() { RegisterApp(&MockApp{definition: nil}) })

	assert.PanicsWithValue(t, "nil AppResolver", func() { RegisterDefaultApp(nil) })
	assert.NotPanics(t, func() { RegisterDefaultApp(&MockAppResolver{}) })
	assert.PanicsWithValue(t, "default resolver already set", func() {
		RegisterDefaultApp(&MockAppResolver{})
	})

	assert.Panics(t, func() { Resolve(nil) })
}

type defaultRes struct{ def wallet.Address }

func (r defaultRes) Resolve(wallet.Address) (App, error) {
	return NewMockApp(r.def), nil
}

func testAppRegistryIdentity(t *testing.T, rng *rand.Rand) {
	resetAppRegistry()
	a0 := newRandomMockApp(rng)
	RegisterApp(a0)
	assertIdentity(t, a0)

	a1 := newRandomMockApp(rng)
	RegisterAppResolver(a1.Def().Equals, &MockAppResolver{})
	assertIdentity(t, a1)

	a2 := newRandomMockApp(rng)
	RegisterDefaultApp(defaultRes{a2.Def()})
	assertIdentity(t, a2)
}

func assertIdentity(t *testing.T, expected App) {
	actual, err := Resolve(expected.Def())
	assert.NoError(t, err)
	assert.True(t, actual.Def().Equals(expected.Def()))
}

func newRandomMockApp(rng *rand.Rand) App {
	return NewMockApp(wtest.NewRandomAddress(rng))
}

func resetAppRegistry() {
	appRegistry.Lock()
	defer appRegistry.Unlock()
	appRegistry.resolvers = nil
	appRegistry.singles = make(map[wallet.AddrKey]App)
	appRegistry.defaultRes = nil
}
