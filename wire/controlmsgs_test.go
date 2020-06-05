// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"testing"
)

func TestPingMsg(t *testing.T) {
	TestMsg(t, NewPingMsg())
}

func TestPongMsg(t *testing.T) {
	TestMsg(t, NewPongMsg())
}

func TestShutdownMsg(t *testing.T) {
	TestMsg(t, &ShutdownMsg{"m2384ordkln fb30954390582"})
}
