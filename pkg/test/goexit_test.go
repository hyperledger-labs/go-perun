// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckGoexit(t *testing.T) {
	assert.True(t, CheckGoexit(runtime.Goexit))
	assert.False(t, CheckGoexit(func() { panic("") }))
	assert.False(t, CheckGoexit(func() { panic(nil) }))
	assert.False(t, CheckGoexit(func() {}))
}
