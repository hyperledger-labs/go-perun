// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer // import "perun.network/go-perun/peer"

import (
	wire "perun.network/go-perun/wire/msg"
)

func init() {
	wire.RegisterPeerDecode(decodeMsg)
	wire.RegisterPeerEncode(encodeMsg)
}
