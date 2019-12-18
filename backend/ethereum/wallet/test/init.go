// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test // import "perun.network/go-perun/backend/ethereum/wallet/test"

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"perun.network/go-perun/wallet/test"
)

var ks *keystore.KeyStore

func init() {
	rnd := newRandomizer()
	ks = rnd.wallet.Ks
	test.SetRandomizer(rnd)
}

// GetKeystore returns the currently set keystore for tests.
func GetKeystore() *keystore.KeyStore {
	return ks
}
