// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client

import (
	"math/rand"

	"github.com/sirupsen/logrus"

	"perun.network/go-perun/apps/payment"
	_ "perun.network/go-perun/backend/sim/channel" // backend init
	_ "perun.network/go-perun/backend/sim/wallet"  // backend init
	plogrus "perun.network/go-perun/log/logrus"
	wallettest "perun.network/go-perun/wallet/test"
)

// This file initializes the blockchain and logging backend for both, whitebox
// and blackbox tests (files *_test.go in packages client and client_test).
func init() {
	plogrus.Set(logrus.WarnLevel, &logrus.TextFormatter{ForceColors: true})

	// Tests of package client use the payment app for now...
	rng := rand.New(rand.NewSource(0x9a09c3e008f1242))
	appDef := wallettest.NewRandomAddress(rng)
	payment.SetAppDef(appDef) // payment app address has to be set once at startup
}
