// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import (
	"testing"
)

func TestPingMsg(t *testing.T) {
	TestMsg(t, NewPingMsg())
}

func TestPongMsg(t *testing.T) {
	TestMsg(t, NewPongMsg())
}
