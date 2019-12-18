// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package peer

import (
	"net"
)

// newPipeConnPair creates endpoints that are connected via pipes.
func newPipeConnPair() (a Conn, b Conn) {
	c0, c1 := net.Pipe()
	return NewIoConn(c0), NewIoConn(c1)
}
