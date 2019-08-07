// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package channel holds the core channel data structures.
// Those data structures are interpreted by the adjudicator.
package channel // import "perun.network/go-perun/channel"

import (
	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/io"
)

type Sig = []byte

// State is the full state of a state channel (app).
// It does not include the channel parameters. However, the included channel ID
// should be a commitment to the channel parameters by hash-digesting them.
// The state is the piece of data that is signed and sent to the adjudicator
// during disputes.
type State struct {
	// id is the immutable id of the channel this state belongs to
	ID ID
	// version counter
	Version uint64
	// Allocation is the current allocation of channel assets to
	// the channel participants and apps running inside this channel.
	Allocation
	// Data is the app data. The App of Params works with this object.
	Data io.Serializable
	// IsFinal
	IsFinal bool
}

// newState creates a new state, checking that the parameters and allocation are
// compatible. This function is not exported as a user of the channel package
// would usually not need to create a State directly, but creates a Machine
// instead.
func newState(params *Params, initBals Allocation, initData io.Serializable) (*State, error) {
	// sanity checks
	n := len(params.Parts)
	if n != len(initBals.OfParts) {
		return nil, errors.New("number of participants in parameters and initial balances don't match.")
	}

	return &State{
		ID:         params.ID(),
		Version:    0,
		Allocation: initBals,
		Data:       initData,
	}, nil
}

// Transaction is a channel state together with valid signatures from the
// channel participants.
type Transaction struct {
	*State
	Sigs []Sig
}
