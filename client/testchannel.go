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
	"context"

	"perun.network/go-perun/channel"
)

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

// Register exposes dispute registration.
func (c *TestChannel) Register(ctx context.Context) error {
	return c.registerDispute(ctx)
}
