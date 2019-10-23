// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/wallet/test"

import (
	"math/rand"

	"perun.network/go-perun/wallet"
)

// Backend is the wallet testing backend. It currently supports the generation
// of random addresses.
type Backend interface {
	NewRandomAddress(*rand.Rand) wallet.Address
}

var backend Backend

// SetBackend sets the wallet testing backend. It may be set multiple times.
func SetBackend(b Backend) {
	backend = b
}

// NewRandomAddress returns a new random address by calling the currently set
// wallet testing backend.
func NewRandomAddress(rng *rand.Rand) wallet.Address {
	return backend.NewRandomAddress(rng)
}
