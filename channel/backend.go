// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

// backend is set to the global channel backend. It must be set through
// backend.Set(Collection).
var backend Backend

type Backend interface {
	// ChannelID infers the channel id of a channel from its parameters. Usually,
	// this should be a hash digest of some or all fields of the parameters.
	ChannelID(*Params) ID
}

// SetBackend sets the global channel backend. Must not be called directly but
// through backend.Set().
func SetBackend(b Backend) {
	backend = b
}
