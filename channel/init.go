// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel // import "perun.network/go-perun/channel"

import (
	wire "perun.network/go-perun/wire/msg"
)

func init() {
	wire.RegisterDecoder(wire.Channel, decodeMsg)
	wire.RegisterEncoder(wire.Channel, encodeMsg)

	wire.RegisterDecoder(wire.Peer, decodePeerMsg)
	wire.RegisterEncoder(wire.Peer, encodePeerMsg)
}
