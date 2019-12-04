// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestAlreadySettledError(t *testing.T) {
	assert := assert.New(t)
	err := NewAlreadySettledError(42, 123)
	perr, ok := errors.Cause(err).(*AlreadySettledError)
	assert.True(ok)
	assert.Equal(Index(42), perr.PeerIdx)
	assert.Equal(uint64(123), perr.Version)
	assert.True(IsAlreadySettledError(err))
	assert.True(IsAlreadySettledError(perr))
	assert.False(IsAlreadySettledError(errors.New("no AlreadySettledError")))
}
