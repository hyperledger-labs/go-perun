// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateResponder_Accept_NilArgs(t *testing.T) {
	err := new(UpdateResponder).Accept(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestUpdateResponder_Reject_NilArgs(t *testing.T) {
	err := new(UpdateResponder).Reject(nil, "reason")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

// Channel.Update() is defined in `client/update.go` so its test can be found
// here as well
func TestChannel_Update_NilArgs(t *testing.T) {
	err := new(Channel).Update(nil, *new(ChannelUpdate))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}
