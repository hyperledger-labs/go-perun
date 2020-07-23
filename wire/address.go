// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wire

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
