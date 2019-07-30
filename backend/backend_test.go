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

	assert.NotPanics(t, func() { Set(nilBackend) }, "First backend.Set should not panic")
	assert.Panics(t, func() { Set(nilBackend) }, "Second backend.Set should panic")
}
