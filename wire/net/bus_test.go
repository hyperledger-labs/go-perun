// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package net_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/net"
	nettest "perun.network/go-perun/wire/net/test"
	wiretest "perun.network/go-perun/wire/test"
)

func TestBus(t *testing.T) {
	const numClients = 16
	const numMsgs = 16

	var hub nettest.ConnHub

	wiretest.GenericBusTest(t, func(acc wire.Account) wire.Bus {
		bus := net.NewBus(acc, hub.NewNetDialer())
		hub.OnClose(func() { bus.Close() })
		go bus.Listen(hub.NewNetListener(acc.Address()))
		return bus
	}, numClients, numMsgs)

	assert.NoError(t, hub.Close())
}
