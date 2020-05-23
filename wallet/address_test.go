// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package wallet_test

import (
	"io"
	"math/rand"
	"testing"

	_ "perun.network/go-perun/backend/sim/wallet"
	iotest "perun.network/go-perun/pkg/io/test"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
)

type testAddresses struct {
	addrs wallet.AddressesWithLen
}

func (t *testAddresses) Encode(w io.Writer) error {
	return t.addrs.Encode(w)
}

func (t *testAddresses) Decode(r io.Reader) error {
	return t.addrs.Decode(r)
}

func TestAddresses_Serializer(t *testing.T) {
	rng := rand.New(rand.NewSource(0xC00FED))

	addrs := wallettest.NewRandomAddresses(rng, 0)
	iotest.GenericSerializerTest(t, &testAddresses{addrs})

	addrs = wallettest.NewRandomAddresses(rng, 1)
	iotest.GenericSerializerTest(t, &testAddresses{addrs})

	addrs = wallettest.NewRandomAddresses(rng, 5)
	iotest.GenericSerializerTest(t, &testAddresses{addrs})
}
