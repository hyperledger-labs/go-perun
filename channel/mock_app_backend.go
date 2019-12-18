// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/channel"

import (
	"perun.network/go-perun/wallet"
)

type MockAppBackend struct{}

var _ AppBackend = &MockAppBackend{}

func (MockAppBackend) AppFromDefinition(addr wallet.Address) (App, error) {
	return NewMockApp(addr), nil
}
