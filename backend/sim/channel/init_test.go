// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

// +build !wrap_test

package channel

import (
	"testing"

	"perun.network/go-perun/channel"
)

func TestSetBackend(t *testing.T) {
	channel.SetBackendTest(t)
}
