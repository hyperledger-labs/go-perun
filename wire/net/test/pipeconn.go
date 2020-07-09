// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"net"

	wirenet "perun.network/go-perun/wire/net"
)

// NewPipeConnPair creates endpoints that are connected via pipes.
func NewPipeConnPair() (a wirenet.Conn, b wirenet.Conn) {
	c0, c1 := net.Pipe()
	return wirenet.NewIoConn(c0), wirenet.NewIoConn(c1)
}
