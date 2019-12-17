// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package wallet

// Account represents a single account.
type Account interface {
	// Address used by this account.
	Address() Address

	// SignData requests a signature from this account.
	// It returns the signature or an error.
	SignData(data []byte) ([]byte, error)
}
