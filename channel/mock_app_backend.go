// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package channel // import "perun.network/go-perun/channel"

import (
	"perun.network/go-perun/wallet"
)

// MockAppBackend is the backend for a mock app.
type MockAppBackend struct{}

var _ AppBackend = &MockAppBackend{}

// AppFromDefinition creates a new MockApp with the provided address.
func (MockAppBackend) AppFromDefinition(addr wallet.Address) (App, error) {
	return NewMockApp(addr), nil
}
