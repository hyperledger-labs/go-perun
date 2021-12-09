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

package ethereum_test

import (
	"testing"

	pkgtest "polycry.pt/poly-go/test"

	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/wire/test"
)

func Test_Asset_GenericMarshaler(t *testing.T) {
	rng := pkgtest.Prng(t)
	for i := 0; i < 10; i++ {
		asset := ethwallettest.NewRandomAddress(rng)
		test.GenericMarshalerTest(t, &asset)
	}
}
