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
// It should only be created through NewParams()
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

// NewParams creates Params from the given data and performs sanity checks. The
// channel id is also calculated here and persisted because it probably is an
// expensive hash operation.
func NewParams(
	challengeDuration uint64,
	parts []wallet.Address,
	app App,
	nonce *big.Int,
) (*Params, error) {
	if challengeDuration == 0 {
		return nil, errors.New("ChallengeDuration must be > 0")
	}
	if len(parts) < 2 {
		return nil, errors.New("need at least two participants")
	}
	if nonce == nil {
		return nil, errors.New("nonce must not be nil")
	}
	if !IsStateApp(app) && !IsActionApp(app) {
		return nil, errors.New("app must either be a StateApp or ActionApp")
	}

	p := &Params{
		ChallengeDuration: challengeDuration,
		Parts:             parts,
		App:               app,
		Nonce:             nonce,
	}
	// probably an expensive hash operation, do it only once during creation.
	p.id = ChannelID(p)

	return p, nil
}
