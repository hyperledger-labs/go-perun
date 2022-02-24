// Copyright 2022 - See NOTICE file for copyright holders.
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

package wallet_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"polycry.pt/poly-go/test"

	"perun.network/go-perun/backend/sim/wallet"
)

func TestAddressJSONMarshaling(t *testing.T) {
	rng := test.Prng(t)
	addr := wallet.NewRandomAddress(rng)

	b, err := json.Marshal(addr)
	require.NoError(t, err)

	addr1 := new(wallet.Address)
	require.NoError(t, json.Unmarshal(b, addr1))
	require.Equal(t, addr, addr1)
}
