// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"
	perun "perun.network/go-perun/wallet"
)

// ZeroAddr is the constant address 0x0000000000000000000000000000000000000000
var ZeroAddr = common.BytesToAddress(make([]byte, 20, 20))

// Address represents an ethereum address
type Address struct {
	common.Address
}

// Bytes converts this address to bytes
func (a *Address) Bytes() []byte {
	return a.Address.Bytes()
}

// String converts this address to a string
func (a *Address) String() string {
	return a.Address.String()
}

// Equals checks the equality of two addresses
func (a *Address) Equals(addr perun.Address) bool {
	return bytes.Equal(a.Address.Bytes(), addr.Bytes())
}
