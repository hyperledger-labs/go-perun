// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

// +build !wrap_test

package channel

import (
	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
)

func init() {
	channel.SetBackend(new(backend))
	test.SetRandomizer(new(randomizer))
}
