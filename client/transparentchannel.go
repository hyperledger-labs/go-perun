// Copyright 2023 - See NOTICE file for copyright holders.
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

import "perun.network/go-perun/channel"

// TransparentChannel grants access to the `SignedState` of the underlying
// channel, which is otherwise hidden by `Channel`. Behaves like a normal channel in
// all other cases.
type TransparentChannel struct {
	*Channel
}

// NewTransparentChannel returns a new `TransparentChannel` from a `Channel`.
func NewTransparentChannel(c *Channel) *TransparentChannel {
	return &TransparentChannel{c}
}

// SignedState returns the current signed state of the channel.
func (c *TransparentChannel) SignedState() channel.SignedState {
	return channel.SignedState{
		Params: c.Params(),
		State:  c.machine.CurrentTX().State,
		Sigs:   c.machine.CurrentTX().Sigs,
	}
}
