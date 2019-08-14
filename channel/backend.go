// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import "perun.network/go-perun/wallet"

// backend is set to the global channel backend. It must be set through
// backend.Set(Collection).
var backend Backend

type Backend interface {
	// ChannelID infers the channel id of a channel from its parameters. Usually,
	// this should be a hash digest of some or all fields of the parameters.
	// If any parameters are omitted from the ChannelID digest, they need to be
	// signed together with the State in Sign().
	ChannelID(*Params) ID

	// Sign signs a channel's State with the given Account. Returns the signature
	// or an error and a nil signature, if not successful.
	Sign(wallet.Account, *Params, *State) (Sig, error)

	// Verify verifies that the provided signature on the state belongs to the
	// provided address.
	Verify(addr wallet.Address, params *Params, state *State, sig Sig) (bool, error)
}

// SetBackend sets the global channel backend. Must not be called directly but
// through backend.Set().
func SetBackend(b Backend) {
	backend = b
}

func ChannelID(p *Params) ID {
	return backend.ChannelID(p)
}

func Sign(a wallet.Account, p *Params, s *State) (Sig, error) {
	return backend.Sign(a, p, s)
}

func Verify(addr wallet.Address, params *Params, state *State, sig Sig) (bool, error) {
	return backend.Verify(addr, params, state, sig)
}
