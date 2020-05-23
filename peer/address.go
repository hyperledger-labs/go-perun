// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package peer

import (
	stdio "io"

	"perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
)

var _ io.Serializer = (*Addresses)(nil)
var _ io.Serializer = (*AddressesWithLen)(nil)

// Address is a Perun node's public Perun address, which is used as a permanent
// identity within the Perun peer-to-peer network. For now, it is just a stub.
type Address = wallet.Address

// Addresses is a helper type for encoding and decoding address slices in
// situations where the length of the slice is known.
type Addresses = wallet.Addresses

// AddressesWithLen is a helper type for encoding and decoding address slices
// of unknown length.
type AddressesWithLen = wallet.AddressesWithLen

// DecodeAddress decodes a peer address.
func DecodeAddress(r stdio.Reader) (Address, error) {
	return wallet.DecodeAddress(r)
}
