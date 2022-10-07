// Copyright 2021 - See NOTICE file for copyright holders.
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

	"perun.network/go-perun/backend/sim/channel"
	"perun.network/go-perun/wire/test"
	pkgtest "polycry.pt/poly-go/test"
)

func Test_Asset_GenericMarshaler(t *testing.T) {
	rng := pkgtest.Prng(t)
	for n := 0; n < 10; n++ {
		test.GenericMarshalerTest(t, &channel.Asset{ID: rng.Int63()})
	}
}
