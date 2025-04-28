// Copyright 2025 - See NOTICE file for copyright holders.
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

package libp2p_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/net/libp2p"
	"perun.network/go-perun/wire/test"

	pkgtest "polycry.pt/poly-go/test"
)

func TestAddress(t *testing.T) {
	test.TestAddressImplementation(t, func() wire.Address {
		return libp2p.NewAddress("")
	}, func(rng *rand.Rand) wire.Address {
		return libp2p.NewRandomAddress(rng)
	})
}

func TestSignature(t *testing.T) {
	rng := pkgtest.Prng(t)
	acc := libp2p.NewRandomAccount(rng)
	assert.NotNil(t, acc)
	defer acc.Close()

	msg := []byte("test message")
	sig, err := acc.Sign(msg)
	assert.NoError(t, err)

	addr := acc.Address()
	assert.NoError(t, addr.Verify(msg, sig))
}
