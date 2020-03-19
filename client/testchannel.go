// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client

import "perun.network/go-perun/channel"

// TestChannel grants access to the `AdjudicatorRequest` which is otherwise
// hidden by `Channel`. Behaves like a `Channel` in all other cases.
//
// Only used for testing.
type TestChannel struct {
	*Channel
}

// NewTestChannel creates a new `TestChannel` from a `Channel`.
func NewTestChannel(c *Channel) *TestChannel {
	return &TestChannel{c}
}

// AdjudicatorReq returns the `AdjudicatorReq` of the underlying machine.
func (c *TestChannel) AdjudicatorReq() channel.AdjudicatorReq {
	return c.machine.AdjudicatorReq()
}
