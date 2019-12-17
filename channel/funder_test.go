// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestPeerTimedOutFundingError(t *testing.T) {
	assert := assert.New(t)
	err := NewPeerTimedOutFundingError(42)
	perr, ok := errors.Cause(err).(*PeerTimedOutFundingError)
	assert.True(ok)
	assert.Equal(Index(42), perr.TimedOutPeerIdx)
	assert.True(IsPeerTimedOutFundingError(err))
	assert.True(IsPeerTimedOutFundingError(perr))
	assert.False(IsPeerTimedOutFundingError(errors.New("no PeerTimedOutFundingError")))
}
