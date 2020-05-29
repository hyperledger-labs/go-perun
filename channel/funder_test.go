// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package channel

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetFundingError(t *testing.T) {
	assert := assert.New(t)
	err := &AssetFundingError{42, []Index{1, 2, 3, 4}}
	perr, ok := errors.Cause(err).(*AssetFundingError)
	assert.True(ok)
	assert.True(IsAssetFundingError(err))
	assert.True(IsAssetFundingError(perr))
	assert.Equal(42, perr.Asset)
	assert.Equal(Index(1), perr.TimedOutPeers[0])
	assert.Equal(Index(2), perr.TimedOutPeers[1])
	assert.Equal(Index(3), perr.TimedOutPeers[2])
	assert.Equal(Index(4), perr.TimedOutPeers[3])
	assert.Equal(4, len(perr.TimedOutPeers))
	assert.Equal(perr.Error(), "Funding Error on asset [42] peers: [1], [2], [3], [4], did not fund channel in time")
	assert.False(IsAssetFundingError(errors.New("not a asset funding error")))
}

func TestFundingTimeoutError(t *testing.T) {
	assert := assert.New(t)
	errs := []*AssetFundingError{
		{42, []Index{1, 2}},
		{1337, []Index{0, 2}},
		{7531, []Index{1, 3}},
	}
	err := NewFundingTimeoutError(errs)
	perr, ok := errors.Cause(err).(*FundingTimeoutError)
	require.True(t, ok)
	assert.True(IsFundingTimeoutError(err))
	assert.True(IsFundingTimeoutError(perr))
	assert.Equal(42, perr.Errors[0].Asset)
	assert.Equal(Index(1), perr.Errors[0].TimedOutPeers[0])
	assert.Equal(Index(2), perr.Errors[0].TimedOutPeers[1])
	assert.Equal(1337, perr.Errors[1].Asset)
	assert.Equal(Index(0), perr.Errors[1].TimedOutPeers[0])
	assert.Equal(Index(2), perr.Errors[1].TimedOutPeers[1])
	assert.Equal(7531, perr.Errors[2].Asset)
	assert.Equal(Index(1), perr.Errors[2].TimedOutPeers[0])
	assert.Equal(Index(3), perr.Errors[2].TimedOutPeers[1])
	assert.Equal(3, len(perr.Errors))
	assert.Equal(perr.Error(), "Funding Error on asset [42] peers: [1], [2], did not fund channel in time; Funding Error on asset [1337] peers: [0], [2], did not fund channel in time; Funding Error on asset [7531] peers: [1], [3], did not fund channel in time; ")
	// test no funding timeout error
	assert.False(IsFundingTimeoutError(errors.New("no FundingTimeoutError")))
	// nil input should not return error
	assert.NoError(NewFundingTimeoutError(nil))
}
