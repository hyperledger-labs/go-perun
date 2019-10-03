// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/wallet/test"

import (
	"math/rand"

	"perun.network/go-perun/wallet"
)

type Backend interface {
	NewRandomAddress(*rand.Rand) wallet.Address
}

var backend Backend

func SetBackend(b Backend) {
	backend = b
}

func NewRandomAddress(rng *rand.Rand) wallet.Address {
	return backend.NewRandomAddress(rng)
}
