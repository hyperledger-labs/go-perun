// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire_test

import (
	"testing"

	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/test"
)

func TestLocalBus(t *testing.T) {
	bus := wire.NewLocalBus()
	test.GenericBusTest(t, func(wire.Account) wire.Bus {
		return bus
	}, 16, 10)
}
