// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package key_test

import (
	"testing"

	"perun.network/go-perun/db/key"

	"github.com/stretchr/testify/assert"
)

func TestNext(t *testing.T) {
	assert.Equal(t, key.Next(""), "\x00")
	assert.Equal(t, key.Next("a"), "a\x00")
}

func TestIncPrefix(t *testing.T) {
	assert.Equal(t, key.IncPrefix(""), "")
	assert.Equal(t, key.IncPrefix("\x00"), "\x01")
	assert.Equal(t, key.IncPrefix("a"), "b")
	assert.Equal(t, key.IncPrefix("zoo"), "zop")
	assert.Equal(t, key.IncPrefix("\xff"), "")
	assert.Equal(t, key.IncPrefix("\xffa"), "\xffb")
	assert.Equal(t, key.IncPrefix("a\xff"), "b")
	assert.Equal(t, key.IncPrefix("\xff\xff\xff"), "")
}
