// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package channel holds the core channel data structures.
// Those data structures are interpreted by the adjudicator.
package channel // import "perun.network/go-perun/channel"

import (
	"perun.network/go-perun/pkg/io"
)

// State is the full state of a state channel (app).
// It does not include the channel parameters. However, the included channel ID
// should be a commitment to the channel parameters by hash-digesting them.
// The state is the piece of data that is signed and sent to the adjudicator
// during disputes.
type State struct {
	// id is the immutable id of the channel this state belongs to
	id ID
	// Version counter
	Version uint64
	// Allocation is the current allocation of channel assets to
	// the channel participants and apps running inside this channel.
	Allocation
	// Data is the app data. The App of Params works with this object.
	Data io.Serializable
	// IsFinal
	IsFinal bool
}

func (s *State) ID() ID {
	return s.id
}
