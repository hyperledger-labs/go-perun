// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package ethereum // import "perun.network/go-perun/backend/ethereum"

import (
	_ "perun.network/go-perun/backend/ethereum/channel" // backend init
	_ "perun.network/go-perun/backend/ethereum/wallet"  // backend init
)
