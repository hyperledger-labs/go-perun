// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package common provides type abstractions that are used throughout go-perun.
package common

import "encoding"

// Address represents a identifier used in a cryptocurrency.
// It is dependent on the currency and needs to be implemented for every blockchain.
type Address interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// PubKey represents an unabridged public key
// It is dependent on the currency and needs to be implemented for every blockchain.
type PubKey interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler

	// ToAddress converts this public key to an address in the respective cryptocurrency.
	ToAddress() Address

	// VerifySign verifies whether a signature was signed by the corresponding sk to this pk.
	VerifySign(sign []byte) error
}
