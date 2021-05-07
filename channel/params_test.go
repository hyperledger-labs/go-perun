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

package channel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	simwallet "perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/pkg/io"
	iotest "perun.network/go-perun/pkg/io/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func TestParams_Clone(t *testing.T) {
	rng := pkgtest.Prng(t)
	params := test.NewRandomParams(rng)
	pkgtest.VerifyClone(t, params)
}

func TestParams_Serializer(t *testing.T) {
	rng := pkgtest.Prng(t)
	params := make([]io.Serializer, 10)
	for i := range params {
		var p *channel.Params
		if i&1 == 0 {
			p = test.NewRandomParams(rng, test.WithoutApp())
		} else {
			p = test.NewRandomParams(rng)
		}
		params[i] = p
	}

	iotest.GenericSerializerTest(t, params...)
}

func TestValidateParameters(t *testing.T) {
	rng := pkgtest.Prng(t)

	t.Run("valid", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			params := test.NewRandomParams(rng)
			err := channel.ValidateParameters(params.ChallengeDuration, params.Parts, params.App, params.Nonce)
			assert.NoError(t, err)
		}
	})
	t.Run("parts-nil", func(t *testing.T) {
		params := test.NewRandomParams(rng, test.WithNumParts(10))
		params.Parts[5] = nil
		err := channel.ValidateParameters(params.ChallengeDuration, params.Parts, params.App, params.Nonce)
		assert.Error(t, err)
	})
	t.Run("parts-nil-interface", func(t *testing.T) {
		params := test.NewRandomParams(rng)
		params.Parts[0] = (*simwallet.Address)(nil)
		err := channel.ValidateParameters(params.ChallengeDuration, params.Parts, params.App, params.Nonce)
		assert.Error(t, err)
	})
}
