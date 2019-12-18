// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package atomic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	assert := assert.New(t)

	var b Bool
	assert.False(b.IsSet())
	b.Set()
	assert.True(b.IsSet())
	assert.False(b.TrySet())
	assert.True(b.IsSet())

	b.Unset()
	assert.False(b.IsSet())
	assert.True(b.TrySet())
	assert.True(b.IsSet())
	assert.True(b.TryUnset())
	assert.False(b.TryUnset())
	assert.False(b.IsSet())
}
