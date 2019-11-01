// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	wire "perun.network/go-perun/wire/msg"
)

func TestDummyPeerMsg(t *testing.T) {
	wire.TestMsg(t, &DummyPeerMsg{msg{}, int64(-0x7172635445362718)})
}

func TestMsgType_String(t *testing.T) {
	assert.Equal(t, MsgType(100).String(), "100")
	assert.Equal(t, PeerDummy.String(), "DummyPeerMsg")
}

func TestMsgType_Valid(t *testing.T) {
	assert.True(t, MsgType(PeerDummy).Valid())
	assert.False(t, msgTypeEnd.Valid())
}
