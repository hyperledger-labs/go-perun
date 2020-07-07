// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBroadcaster_Send broadcasts a message and check that the message is sent
// correctly.
func TestBroadcaster_Send(t *testing.T) {
	N := 5
	peers := make([]*Endpoint, N)
	msg := NewPingMsg()
	check := func(m Msg) { assert.Equal(t, msg, m) }

	for i := 0; i < N; i++ {
		peers[i] = newEndpoint(nil, newMockConn(check), nil)
	}

	b := NewBroadcaster(peers)

	assert.NoError(t, b.Send(context.Background(), msg), "broadcast must succeed")
}

// TestBroadcaster_Send_Error tests that when a single transmission fails, thee
// whole operation fails.
func TestBroadcaster_Send_Error(t *testing.T) {
	N := 5
	peers := make([]*Endpoint, N)

	for i := 0; i < N; i++ {
		peers[i] = newEndpoint(nil, newMockConn(nil), nil)
	}

	peers[1].Close()

	b := NewBroadcaster(peers)

	err := b.Send(context.Background(), NewPingMsg())
	require.IsType(t, &BroadcastError{}, errors.Cause(err))
	berr := errors.Cause(err).(*BroadcastError)

	require.Error(t, err, "broadcast must fail")
	assert.Equal(t, len(berr.errors), 1)
	assert.Equal(t, berr.errors[0].index, 1)
	assert.Equal(t, err.Error(), "failed to send message:\npeer[1]: "+berr.errors[0].err.Error())
}
