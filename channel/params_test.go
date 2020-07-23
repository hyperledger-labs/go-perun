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
	"math/rand"
	"testing"

	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/pkg/io"
	iotest "perun.network/go-perun/pkg/io/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func TestParams_Clone(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDDDD))
	params := test.NewRandomParams(rng)
	pkgtest.VerifyClone(t, params)
}

func TestParams_Serializer(t *testing.T) {
	rng := rand.New(rand.NewSource(0xC00FED))
	params := make([]io.Serializer, 10)
	for i := range params {
		params[i] = test.NewRandomParams(rng)
	}

	iotest.GenericSerializerTest(t, params...)
}
