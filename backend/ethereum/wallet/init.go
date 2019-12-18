// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package wallet // import "perun.network/go-perun/backend/ethereum/wallet"

import (
	perunwallet "perun.network/go-perun/wallet"
)

func init() {
	perunwallet.SetBackend(new(Backend))
}
