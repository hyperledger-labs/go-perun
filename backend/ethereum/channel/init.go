// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import (
	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
)

// GasLimit is the max amount of gas we want to send per transaction.
const GasLimit = 200000

func init() {
	channel.SetBackend(new(Backend))
	test.SetRandomizer(new(randomizer))
}
