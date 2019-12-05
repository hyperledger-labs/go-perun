// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	nilBackend := Collection{
		Channel: nil,
		Wallet:  nil,
	}

	// We cannot test here, that it is set to non nil, since it is not part of this package.
	// We should kick out Collection.
	assert.Panics(t, func() { Set(nilBackend) }, "setting a backend twice should panic")
}
