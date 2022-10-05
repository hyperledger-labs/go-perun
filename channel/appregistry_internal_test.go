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
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"polycry.pt/poly-go/test"
)

func TestAppRegistry(t *testing.T) {
	test.OnlyOnce(t)
	rng := test.Prng(t)

	backup := struct {
		resolvers  []appRegEntry
		singles    map[AppIDKey]App
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
	t.Helper()
	resetAppRegistry()
	assert.Panics(t, func() { RegisterAppResolver(nil, nil) })
	assert.Panics(t, func() { RegisterAppResolver(func(AppID) bool { return true }, nil) })
	assert.Panics(t, func() { RegisterAppResolver(nil, &MockAppResolver{}) })

	assert.Panics(t, func() { RegisterApp(nil) })
	assert.Panics(t, func() { RegisterApp(&MockApp{definition: nil}) })

	assert.PanicsWithValue(t, "nil AppResolver", func() { RegisterDefaultApp(nil) })
	assert.NotPanics(t, func() { RegisterDefaultApp(&MockAppResolver{}) })
	assert.PanicsWithValue(t, "default resolver already set", func() {
		RegisterDefaultApp(&MockAppResolver{})
	})

	assert.Panics(t, func() { Resolve(nil) }) //nolint:errcheck
}

type defaultRes struct{ def AppID }

func (r defaultRes) Resolve(AppID) (App, error) {
	return NewMockApp(r.def), nil
}

func testAppRegistryIdentity(t *testing.T, rng *rand.Rand) {
	t.Helper()
	resetAppRegistry()
	a0 := newRandomMockApp(rng)
	RegisterApp(a0)
	assertIdentity(t, a0)

	a1 := newRandomMockApp(rng)
	RegisterAppResolver(a1.Def().Equal, &MockAppResolver{})
	assertIdentity(t, a1)

	a2 := newRandomMockApp(rng)
	RegisterDefaultApp(defaultRes{a2.Def()})
	assertIdentity(t, a2)
}

func assertIdentity(t *testing.T, expected App) {
	t.Helper()
	actual, err := Resolve(expected.Def())
	assert.NoError(t, err)
	assert.True(t, actual.Def().Equal(expected.Def()))
}

func newRandomMockApp(rng *rand.Rand) App {
	return NewMockApp(newRandomAppID(rng))
}

func resetAppRegistry() {
	appRegistry.Lock()
	defer appRegistry.Unlock()
	appRegistry.resolvers = nil
	appRegistry.singles = make(map[AppIDKey]App)
	appRegistry.defaultRes = nil
}

func newRandomAppID(rng *rand.Rand) AppID {
	id := appID{}
	rng.Read(id[:])
	return id
}

const appIDLength = 32

type appID [appIDLength]byte

func (id appID) MarshalBinary() (data []byte, err error) {
	return id[:], nil
}

func (id appID) UnmarshalBinary(data []byte) error {
	l := len(data)
	if l != appIDLength {
		return fmt.Errorf("invalid length: %v", l)
	}
	copy(id[:], data)
	return nil
}

func (id appID) Equal(b AppID) bool {
	bTyped, ok := b.(appID)
	return ok && bytes.Equal(id[:], bTyped[:])
}

// Key returns the object key which can be used as a map key.
func (id appID) Key() AppIDKey {
	b, err := id.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return AppIDKey(b)
}
