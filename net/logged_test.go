// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package net_test

import (
	"testing"
	 "net"

	"perun.network/go-perun/net/test"
	_net "perun.network/go-perun/net"
)

func TestWrappedTCPConn(t *testing.T) {
	const address = "localhost:12345"
	const protocol = "tcp"

	listener := func() (net.Listener, error) {
		return _net.Listen(protocol, address)
	}

	dialer := func() (net.Conn, error) {
		return _net.Dial(protocol, address)
	}

	s := &test.Setup{
		ListenerFactory: listener,
		Dialer:   dialer,
	}

	test.GenericListenerTest(t, s)
	test.GenericDoubleConnectTest(t, s)
}
