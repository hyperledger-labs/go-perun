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

package payment

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "perun.network/go-perun/backend/sim" // backend init
	"perun.network/go-perun/wallet/test"
)

func TestRandomizer(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	if backend.def == nil {
		SetAppDef(test.NewRandomAddress(rng))
		// Reset app def during cleanup in case this test runs before TestBackend,
		// which assumes the app def to not be set yet.
		t.Cleanup(func() { backend.def = nil })
	}

	r := new(Randomizer)
	app := r.NewRandomApp(rng)
	assert.True(t, app.Def().Equals(AppDef()))
	assert.IsType(t, &NoData{}, r.NewRandomData(rng))
}
