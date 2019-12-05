// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"io"

	"perun.network/go-perun/wallet"
)

// backend is set to the global channel backend. It must be set through
// backend.Set(Collection).
var backend Backend

type Backend interface {
	// ChannelID infers the channel id of a channel from its parameters. Usually,
	// this should be a hash digest of some or all fields of the parameters.
	// In order to guarantee non-malleability of States, any parameters omitted
	// from the ChannelID digest need to be signed together with the State in
	// Sign().
	ChannelID(*Params) ID

	// Sign signs a channel's State with the given Account.
	// Returns the signature or an error.
	// The framework guarantees to not pass nil Account, *Params or *State, that
	// the IDs of them match and that Params.ID = ChannelID(Params).
	Sign(wallet.Account, *Params, *State) (wallet.Sig, error)

	// Verify verifies that the provided signature on the state belongs to the
	// provided address.
	// The framework guarantees to not pass nil Address, *Params or *State, that
	// the IDs of them match and that Params.ID = ChannelID(Params).
	Verify(addr wallet.Address, params *Params, state *State, sig wallet.Sig) (bool, error)

	// DecodeAsset decodes an asset from a stream.
	DecodeAsset(io.Reader) (Asset, error)
}

// SetBackend sets the global channel backend. Must not be called directly but
// through backend.Set().
func SetBackend(b Backend) {
	if backend != nil || b == nil {
		panic("channel backend already set or nil argument")
	}
	backend = b
}

func ChannelID(p *Params) ID {
	return backend.ChannelID(p)
}

func Sign(a wallet.Account, p *Params, s *State) (wallet.Sig, error) {
	return backend.Sign(a, p, s)
}

func Verify(addr wallet.Address, params *Params, state *State, sig wallet.Sig) (bool, error) {
	return backend.Verify(addr, params, state, sig)
}

func DecodeAsset(r io.Reader) (Asset, error) {
	return backend.DecodeAsset(r)
}
