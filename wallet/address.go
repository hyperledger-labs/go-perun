// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"fmt"

	"perun.network/go-perun/pkg/io"
)

// Address represents a identifier used in a cryptocurrency.
// It is dependent on the currency and needs to be implemented for every blockchain.
type Address interface {
	io.Serializable
	// Bytes should return the representation of the address as byte slice.
	Bytes() []byte
	// String converts this address to a string
	fmt.Stringer
	// Equals checks the equality of two addresses
	Equals(Address) bool
}
