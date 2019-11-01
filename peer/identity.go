// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"perun.network/go-perun/wallet"
)

// Identity is a node's permanent Perun identity, which is used to establish
// authenticity within the Perun peer-to-peer network. For now, it is just a
// stub.
type Identity = wallet.Account
