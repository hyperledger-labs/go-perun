// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"testing"

	_ "perun.network/go-perun/backend/ethereum/wallet/test" // random init
	pkgtest "perun.network/go-perun/pkg/test"
	wallettest "perun.network/go-perun/wallet/test"
)

func TestAuthResponseMsg(t *testing.T) {
	rng := pkgtest.Prng(t)
	TestMsg(t, NewAuthResponseMsg(wallettest.NewRandomAccount(rng)))
}
