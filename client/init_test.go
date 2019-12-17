// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package client

import (
	"github.com/sirupsen/logrus"

	_ "perun.network/go-perun/backend/sim/channel" // backend init
	_ "perun.network/go-perun/backend/sim/wallet"  // backend init
	plogrus "perun.network/go-perun/log/logrus"
)

// This file initializes the blockchain and logging backend for both, whitebox
// and blackbox tests (files *_test.go in packages client and client_test).
func init() {
	plogrus.Set(logrus.TraceLevel, &logrus.TextFormatter{ForceColors: true})
}
