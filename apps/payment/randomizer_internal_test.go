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
	"testing"

	"github.com/stretchr/testify/assert"

	_ "perun.network/go-perun/backend/sim" // backend init
	"perun.network/go-perun/channel"
	pkgtest "polycry.pt/poly-go/test"
)

func TestRandomizer(t *testing.T) {
	rng := pkgtest.Prng(t)

	r := new(Randomizer)
	app := r.NewRandomApp(rng)
	channel.RegisterApp(app)
	regApp, err := channel.Resolve(app.Def())
	assert.NoError(t, err)
	assert.True(t, app.Def().Equal(regApp.Def()))
	assert.True(t, IsData(r.NewRandomData(rng)))
}
