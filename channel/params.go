// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"math/big"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/io"
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
	// Adjudicator resolves channel disputes
	Adjudicator wallet.Address
	// ChallengeDuration in seconds during disputes
	ChallengeDuration uint64
	// Assets are the asset types held in this channel
	Assets []io.Serializable
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
	b Backend,
	adjudicator wallet.Address,
	challengeDuration uint64,
	assets []io.Serializable,
	parts []wallet.Address,
	nonce *big.Int,
) *Params {
	p := &Params{
		Adjudicator:       adjudicator,
		ChallengeDuration: challengeDuration,
		Assets:            assets,
		Parts:             parts,
		Nonce:             nonce,
	}
	// probably an expensive hash operation, do it only once during creation.
	p.id = b.ChannelID(p)

	return p
}

func (p *Params) ValidTransition(from, to *State) (bool, error) {
	if from.id != p.id || to.id != p.id {
		return false, errors.New("states' IDs don't match parameters")
	}

	if from.IsFinal == true {
		return false, newTransitionError(p.id, "cannot advance final state")
	}

	if from.Version+1 != to.Version {
		return false, newTransitionError(p.id, "version must increase by one")
	}

	eq, err := equalSum(from.Allocation, to.Allocation)
	if err != nil {
		return false, err
	}
	if !eq {
		return false, newTransitionError(p.id, "allocations must be preserved.")
	}

	valid, err := p.App.ValidTransition(p, from, to)
	if !valid {
		return false, newTransitionError(p.id, "no valid application state transition")
	}

	return true, nil
}
