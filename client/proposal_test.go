// Copyright 2020 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProposalResponder_Accept_Nil(t *testing.T) {
	p := new(ProposalResponder)
	_, err := p.Accept(nil, new(LedgerChannelProposalAcc))
	assert.Error(t, err, "context")
}

func TestPeerRejectedProposalError(t *testing.T) {
	reason := "some-random-reason"
	err := newPeerRejectedError("update", reason)
	t.Run("direct_error", func(t *testing.T) {
		peerRejectedProposalError := PeerRejectedError{}
		gotPeerRejectedError := errors.As(err, &peerRejectedProposalError)
		require.True(t, gotPeerRejectedError)
		assert.Equal(t, reason, peerRejectedProposalError.Reason)
		assert.Contains(t, err.Error(), reason)
	})

	t.Run("wrapped_error", func(t *testing.T) {
		wrappedError := errors.WithMessage(err, "some higher level error")
		peerRejectedError := PeerRejectedError{}
		gotPeerRejectedError := errors.As(wrappedError, &peerRejectedError)
		require.True(t, gotPeerRejectedError)
		assert.Equal(t, reason, peerRejectedError.Reason)
		assert.Contains(t, err.Error(), reason)
	})
}
