// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import "fmt"

// Address represents a identifier used in a cryptocurrency.
// It is dependent on the currency and needs to be implemented for every blockchain.
type Address interface {
	// Bytes converts this address to bytes
	Bytes() []byte
	// String converts this address to a string
	fmt.Stringer
	// Equals checks the equality of two addresses
	Equals(Address) bool
}
