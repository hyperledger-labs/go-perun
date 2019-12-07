// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet // import "perun.network/go-perun/backend/ethereum/wallet"

import (
	perunwallet "perun.network/go-perun/wallet"
	"perun.network/go-perun/wallet/test"
)

func init() {
	perunwallet.SetBackend(new(Backend))
	test.SetRandomizer(newRandomizer())
}
