// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"math/big"

	"github.com/pkg/errors"

	"perun.network/go-perun/wallet"
)

const IDLen = 32

type ID = [IDLen]byte

var Zero ID = ID{}

// params are a channel's immutable parameters.  A channel's id is the hash of
// (some of) its parameter, as determined by the backend.  All fields should be
// treated as constant.
type Params struct {
	// ChannelID is the channel ID as calculated by the backend
	id ID
	// ChallengeDuration in seconds during disputes
	ChallengeDuration uint64
	// Parts are the channel participants
	Parts []wallet.Address
	// App identifies the application that this channel is running.
	App App
	// Nonce is a randomness to make the channel id unique
	Nonce *big.Int
}

func (p *Params) ID() ID {
	return p.id
}

func NewParams(
	challengeDuration uint64,
	parts []wallet.Address,
	app App,
	nonce *big.Int,
) *Params {
	p := &Params{
		ChallengeDuration: challengeDuration,
		Parts:             parts,
		App:               app,
		Nonce:             nonce,
	}
	// probably an expensive hash operation, do it only once during creation.
	p.id = ChannelID(p)

	return p
}

func (p *Params) ValidTransition(from, to *State) (bool, error) {
	if from.ID != p.id || to.ID != p.id {
		return false, errors.New("states' IDs don't match parameters")
	}

	if from.IsFinal == true {
		return false, newStateTransitionError(p.id, "cannot advance final state")
	}

	if from.Version+1 != to.Version {
		return false, newStateTransitionError(p.id, "version must increase by one")
	}

	eq, err := equalSum(from.Allocation, to.Allocation)
	if err != nil {
		return false, err
	}
	if !eq {
		return false, newStateTransitionError(p.id, "allocations must be preserved.")
	}

	valid, err := p.App.ValidTransition(p, from, to)
	if !valid {
		if err == nil {
			return false, newStateTransitionError(p.id, "no valid application state transition")
		} else if IsStateTransitionError(err) {
			return false, err
		} else {
			return false, errors.WithMessage(err, "runtime error in application's ValidTransition check")
		}
	}

	return true, nil
}
