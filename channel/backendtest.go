// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SetBackendTest(t *testing.T) {
	assert.Panics(t, func() { SetBackend(nil) }, "nil backend set should panic")
	require.NotNil(t, backend, "backend should be already set by init()")
	assert.Panics(t, func() { SetBackend(backend) }, "setting a backend twice should panic")
}
