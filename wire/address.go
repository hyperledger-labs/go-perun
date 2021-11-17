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

	"perun.network/go-perun/wallet"
	"polycry.pt/poly-go/io"
)

var (
	_ io.Serializer = (*Addresses)(nil)
	_ io.Serializer = (*AddressesWithLen)(nil)
)

// Address is a Perun node's network address, which is used as a permanent
// identity within the Perun peer-to-peer network. For now, it is based on type
// wallet.Address.
type Address interface {
	wallet.Address
}

// Addresses is a helper type for encoding and decoding address slices in
// situations where the length of the slice is known.
type Addresses []Address

// AddressesWithLen is a helper type for encoding and decoding address slices
// of unknown length.
type AddressesWithLen []Address

// DecodeAddress decodes a peer address.
func DecodeAddress(r stdio.Reader) (Address, error) {
	return wallet.DecodeAddress(r)
}

// Encode encodes wire addresses.
func (a Addresses) Encode(w stdio.Writer) error {
	return wallet.Addresses(asWalletAddresses(a)).Encode(w)
}

// Encode encodes wire addresses with length.
func (a AddressesWithLen) Encode(w stdio.Writer) error {
	return wallet.AddressesWithLen(asWalletAddresses(a)).Encode(w)
}

// asWalletAddresses converts wire addresses to wallet addresses.
func asWalletAddresses(a []Address) []wallet.Address {
	b := make([]wallet.Address, len(a))
	for i, x := range a {
		b[i] = x
	}
	return b
}

// Decode decodes wallet addresses.
func (a Addresses) Decode(r stdio.Reader) error {
	b := wallet.Addresses(make([]wallet.Address, len(a)))
	if err := b.Decode(r); err != nil {
		return err
	}
	for i, x := range b {
		a[i] = x
	}
	return nil
}

// Decode decodes a wallet address slice of unknown length.
func (a *AddressesWithLen) Decode(r stdio.Reader) error {
	var b wallet.AddressesWithLen
	if err := b.Decode(r); err != nil {
		return err
	}
	*a = make(AddressesWithLen, len(b))
	for i, x := range b {
		(*a)[i] = x
	}
	return nil
}

// IndexOfAddr returns the index of the given address in the address slice,
// or -1 if it is not part of the slice.
func IndexOfAddr(addrs []Address, addr Address) int {
	for i, a := range addrs {
		if addr.Equal(a) {
			return i
		}
	}

	return -1
}
