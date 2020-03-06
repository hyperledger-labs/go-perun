// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

// Package channel holds the core channel data structures.
// Those data structures are interpreted by the adjudicator.
package channel // import "perun.network/go-perun/channel"

import (
	"io"

	"github.com/pkg/errors"

	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
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
		// App identifies the application that this channel is running.
		// We do not want a deep copy here, since the Apps are just an immutable reference.
		// They are only included in the State to support serialization of the `Data` field.
		App App `cloneable:"shallow"`
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
		Sigs []wallet.Sig
	}

	// Data is the data of the application running in this app channel.
	// Decoding happens with App.DecodeData.
	Data interface {
		perunio.Encoder
		// Clone should return a deep copy of the Data object.
		// It should return nil if the Data object is nil.
		Clone() Data
	}
)

var _ perunio.Serializer = new(State)

// newState creates a new state, checking that the parameters and allocation are
// compatible. This function is not exported because a user of the channel
// package would usually not create a State directly. The user receives the
// initial state from the machine instead.
func newState(params *Params, initBals Allocation, initData Data) (*State, error) {
	// sanity checks
	n := len(params.Parts)
	if n != len(initBals.Balances) {
		return nil, errors.New("number of participants in parameters and initial balances don't match")
	}
	if err := initBals.Valid(); err != nil {
		return nil, err
	}

	return &State{
		ID:         params.ID(),
		Version:    0,
		App:        params.App,
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
	// Shallow copy the app.
	clone.App = s.App
	return &clone
}

// Encode encodes a state into an `io.Writer` or returns an `error`
func (s State) Encode(w io.Writer) error {
	err := wire.Encode(w, s.ID, s.Version, s.Allocation, s.IsFinal, s.App.Def(), s.Data)
	return errors.WithMessage(err, "state encode")
}

// Decode decodes a state from an `io.Reader` or returns an `error`
func (s *State) Decode(r io.Reader) error {
	// Decode ID, Version, Allocation, IsFinal
	if err := wire.Decode(r, &s.ID, &s.Version, &s.Allocation, &s.IsFinal); err != nil {
		return errors.WithMessage(err, "id or version decode")
	}
	// Decode app
	var err error
	def, err := wallet.DecodeAddress(r)
	if err != nil {
		return errors.WithMessage(err, "app definition decode")
	}
	s.App, err = AppFromDefinition(def)
	if err != nil {
		return errors.WithMessage(err, "app from definition")
	}
	// Decode app data
	s.Data, err = s.App.DecodeData(r)
	return errors.WithMessage(err, "app decode data")
}

// Clone returns a deep copy of Transaction
func (t *Transaction) Clone() *Transaction {
	var clonedSigs []wallet.Sig
	if t.Sigs != nil {
		clonedSigs = make([]wallet.Sig, len(t.Sigs))
		for i, sig := range t.Sigs {
			if sig != nil {
				clonedSigs[i] = make(wallet.Sig, len(sig))
				copy(clonedSigs[i], sig)
			}
		}
	}
	return &Transaction{
		State: t.State.Clone(),
		Sigs:  clonedSigs,
	}
}
