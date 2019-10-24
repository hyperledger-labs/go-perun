// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"testing"
)

func TestDefaultWalletBackend_Address(t *testing.T) {
	GenericAddressTest(t, &Setup{
		AddrString: "204b49d0acfecbee86904de3aa37b3d28fc0233561de51fbcd34e4be33ee9d53",
		Backend:    new(DefaultWalletBackend),
	})
}
