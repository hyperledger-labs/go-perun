// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransitionError(t *testing.T) {
	assert.False(t, IsTransitionError(errors.New("NoTransitionError")))
	assert.True(t, IsTransitionError(newTransitionError(Zero, "ATransitionError")))
}
