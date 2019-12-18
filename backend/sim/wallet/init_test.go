// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

// +build !wrap_test

package wallet // import "perun.network/go-perun/backend/sim/wallet"

import (
	"testing"

	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wallet/test"
)

func TestSetBackend(t *testing.T) {
	wallet.SetBackendTest(t)
	test.SetRandomizerTest(t)
}
