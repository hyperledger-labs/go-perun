// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"testing"

	wire "perun.network/go-perun/wire/msg"
)

func TestDummyPeerMsg(t *testing.T) {
	wire.TestMsg(t, &DummyPeerMsg{msg{}, int64(-0x7172635445362718)})
}
