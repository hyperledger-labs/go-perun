// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package channel holds the core channel data structures.
// Those data structures are interpreted by the adjudicator.
package channel // import "perun.network/go-perun/channel"

import (
	"io"

	"github.com/pkg/errors"

	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wire"
)

type (
	// State is the full state of a state channel (app).
	// It does not include the channel parameters. However, the included channel ID
	// should be a commitment to the channel parameters by hash-digesting them.
	// The state is the piece of data that is signed and sent to the adjudicator
	// during disputes.
	State struct {
		// id is the immutable id of the channel this state belongs to
		ID ID
		// version counter
		Version uint64
		// Allocation is the current allocation of channel assets to
		// the channel participants and apps running inside this channel.
		Allocation
		// Data is the app-specific data.
		Data Data
		// IsFinal indicates that the channel is in its final state. Such a state
		// can immediately be settled on the blockchain or a funding channel, in
		// case of sub- or virtual channels.
		// A final state cannot be further progressed.
		IsFinal bool
	}

	// Transaction is a channel state together with valid signatures from the
	// channel participants.
	Transaction struct {
		*State
		Sigs []Sig
	}

	// Sig is a single signature
	Sig = []byte

	// Data is the data of the application running in this app channel
	Data interface {
		perunio.Serializable
		// Clone should return a deep copy of the Data object.
		// It should return nil if the Data object is nil.
		Clone() Data
	}

	DummyData struct {
		X uint32
	}
)

// newState creates a new state, checking that the parameters and allocation are
// compatible. This function is not exported because a user of the channel
// package would usually not create a State directly. The user receives the
// initial state from the machine instead.
func newState(params *Params, initBals Allocation, initData Data) (*State, error) {
	// sanity checks
	n := len(params.Parts)
	if n != len(initBals.OfParts) {
		return nil, errors.New("number of participants in parameters and initial balances don't match")
	}
	if err := initBals.valid(); err != nil {
		return nil, err
	}

	return &State{
		ID:         params.ID(),
		Version:    0,
		Allocation: initBals,
		Data:       initData,
	}, nil
}

// Clone makes a deep copy of the State object.
// If it is nil, it returns nil.
// App implementations should use this method when creating the next state from
// an old one.
func (s *State) Clone() *State {
	if s == nil {
		return nil
	}

	clone := *s
	clone.Allocation = s.Allocation.Clone()
	clone.Data = s.Data.Clone()
	return &clone
}

func (d *DummyData) Encode(w io.Writer) error {
	return wire.Encode(w, d.X)
}

func (d *DummyData) Decode(r io.Reader) error {
	return wire.Decode(r, &d.X)
}

func (d *DummyData) Clone() Data {
	return &DummyData{d.X}
}
