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
	"github.com/stretchr/testify/require"

	_ "perun.network/go-perun/backend/sim" // backend init
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestResolver(t *testing.T) {
	pkgtest.OnlyOnce(t)

	rng := pkgtest.Prng(t)
	assert, require := assert.New(t), require.New(t)

	def := test.NewRandomAddress(rng)
	channel.RegisterAppResolver(def.Equal, &Resolver{})

	app, err := channel.Resolve(def)
	assert.NoError(err)
	require.NotNil(app)
	assert.True(def.Equal(app.Def()))
}

func TestData(t *testing.T) {
	assert := assert.New(t)

	assert.NotPanics(func() {
		data := Data()
		assert.Nil(data.Encode(nil))
	})

	assert.NotPanics(func() {
		app := new(App)
		data, err := app.DecodeData(nil)
		assert.NoError(err)
		assert.NotNil(data)
		assert.True(IsData(data))
	})

	data := Data()
	clone := data.Clone()
	assert.IsType(data, clone)
}
