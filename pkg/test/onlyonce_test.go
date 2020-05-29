// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

var onlyOnceTestCalls int32

func TestOnlyOnce1(t *testing.T) {
	testOnlyOnce(t)
}

func TestOnlyOnce2(t *testing.T) {
	testOnlyOnce(t)
}

func testOnlyOnce(t *testing.T) {
	OnlyOnce(t)
	assert.Equal(t, int32(1), atomic.AddInt32(&onlyOnceTestCalls, 1))
}
