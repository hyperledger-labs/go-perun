// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppRandomizerSet(t *testing.T) {
	assert.Panics(t, func() { SetAppRandomizer(nil) }, "nil backend set should panic")
	assert.Panics(t, func() { SetAppRandomizer(&MockAppRandomizer{}) }, "backend should be already set by init()")
}
