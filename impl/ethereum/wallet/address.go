// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"github.com/ethereum/go-ethereum/common"
)

// Address represents an ethereum address
type Address struct {
	common.Address
}

// Bytes converts this address to bytes
(a *Address) Bytes() []byte {
	return a.Address.Bytes()
}

// String converts this address to a string
(a *Address) String() string {
	return a.Address.String()
}

// Equals checks the equality of two addresses
(a *Address) Equals(addr Address) bool {
	return a.Address.Bytes() == addr.Bytes()
}

