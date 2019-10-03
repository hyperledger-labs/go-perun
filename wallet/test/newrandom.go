// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/wallet/test"

import (
	"log"
	"math/rand"

	"perun.network/go-perun/wallet"
)

type TestBackend struct {
	NewRandomAddressFunc func(*rand.Rand) wallet.Address
}

var backend *TestBackend

func SetBackend(new_backend TestBackend) {
	if backend != nil {
		log.Panic("TestBackend set twice")
	}
	backend = &new_backend
}

func NewRandomAddress(rng *rand.Rand) wallet.Address {
	return backend.NewRandomAddressFunc(rng)
}
