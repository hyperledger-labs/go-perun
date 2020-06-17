// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package channel

import (
	"bytes"
	stdio "io"
	"log"
	"math/big"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
)

// IDLen the length of a channelID.
const IDLen = 32

// ID represents a channelID.
type ID = [IDLen]byte

// Zero is the default channelID.
var Zero ID = ID{}

var _ io.Serializer = (*Params)(nil)

// Params are a channel's immutable parameters.  A channel's id is the hash of
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
	App App `cloneable:"shallow"`
	// Nonce is a randomness to make the channel id unique
	Nonce *big.Int
}

// ID returns the channelID of this channel.
func (p *Params) ID() ID {
	return p.id
}

// NewParams creates Params from the given data and performs sanity checks. The
// channel id is also calculated here and persisted because it probably is an
// expensive hash operation.
func NewParams(challengeDuration uint64, parts []wallet.Address, appDef wallet.Address, nonce *big.Int) (*Params, error) {
	if err := ValidateParameters(challengeDuration, len(parts), appDef, nonce); err != nil {
		return nil, errors.WithMessage(err, "invalid parameter for NewParams")
	}
	return NewParamsUnsafe(challengeDuration, parts, appDef, nonce), nil
}

// ValidateParameters checks that the arguments form valid Params:
// * non-zero ChallengeDuration
// * non-nil nonce
// * at least two and at most MaxNumParts parts
// * appDef belongs to either a StateApp or ActionApp
func ValidateParameters(challengeDuration uint64, numParts int, appDef wallet.Address, nonce *big.Int) error {
	if challengeDuration == 0 {
		return errors.New("challengeDuration must be != 0")
	}
	if nonce == nil {
		return errors.New("nonce must not be nil")
	}
	if numParts < 2 {
		return errors.New("need at least two participants")
	}
	if numParts > MaxNumParts {
		return errors.Errorf("too many participants, got: %d max: %d", numParts, MaxNumParts)
	}
	app, err := AppFromDefinition(appDef)
	if err != nil {
		return errors.WithMessage(err, "app from definition")
	}
	if !IsStateApp(app) && !IsActionApp(app) {
		return errors.New("app must be either an Action- or StateApp")
	}
	return nil
}

// NewParamsUnsafe creates Params from the given data and does NOT perform sanity checks.
// The channel id is also calculated here and persisted because it probably is an
// expensive hash operation.
func NewParamsUnsafe(challengeDuration uint64, parts []wallet.Address, appDef wallet.Address, nonce *big.Int) *Params {
	app, err := AppFromDefinition(appDef)
	if err != nil {
		log.Panic("AppFromDefinition on validated parameters returned error")
	}
	p := &Params{
		ChallengeDuration: challengeDuration,
		Parts:             parts,
		App:               app,
		Nonce:             nonce,
	}
	// probably an expensive hash operation, do it only once during creation.
	p.id = CalcID(p)
	return p
}

// Clone returns a deep copy of Params
func (p *Params) Clone() *Params {
	clonedParts := make([]wallet.Address, len(p.Parts))
	for i, v := range p.Parts {
		var buff bytes.Buffer
		v.Encode(&buff)

		addr, err := wallet.DecodeAddress(&buff)
		if err != nil {
			panic("Could not clone params' addresses")
		}
		clonedParts[i] = addr
	}

	return &Params{
		id:                p.ID(),
		ChallengeDuration: p.ChallengeDuration,
		Parts:             clonedParts,
		App:               p.App,
		Nonce:             new(big.Int).Set(p.Nonce)}
}

// Encode uses the pkg/io module to serialize a params instance.
func (p *Params) Encode(w stdio.Writer) error {
	return io.Encode(w,
		p.id,
		p.ChallengeDuration,
		wallet.AddressesWithLen(p.Parts),
		p.App.Def(),
		p.Nonce)
}

// Decode uses the pkg/io module to deserialize a params instance.
func (p *Params) Decode(r stdio.Reader) error {
	var appDef wallet.Address
	err := io.Decode(r,
		&p.id,
		&p.ChallengeDuration,
		(*wallet.AddressesWithLen)(&p.Parts),
		wallet.AddressDec{Addr: &appDef},
		&p.Nonce)
	if err != nil {
		return errors.WithMessage(err, "decode fields")
	}

	p.App, err = AppFromDefinition(appDef)
	return errors.WithMessage(err, "resolve app")
}
