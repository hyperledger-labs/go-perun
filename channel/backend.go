// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package channel

import (
	"perun.network/go-perun/wallet"
)

// backend is set to the global channel backend. Must not be set directly but
// through importing the needed backend.
var backend Backend

// Backend is an interface that needs to be implemented for every blockchain.
// It provides basic functionalities to the framework.
type Backend interface {
	// CalcID infers the channel id of a channel from its parameters. Usually,
	// this should be a hash digest of some or all fields of the parameters.
	// In order to guarantee non-malleability of States, any parameters omitted
	// from the CalcID digest need to be signed together with the State in
	// Sign().
	CalcID(params *Params) ID

	// Sign signs a channel's State with the given Account.
	// Returns the signature or an error.
	Sign(acc wallet.Account, state *State) (wallet.Sig, error)

	// Verify verifies that the provided signature on the state belongs to the
	// provided address.
	Verify(addr wallet.Address, state *State, sig wallet.Sig) (bool, error)

	// NewAsset returns a variable of type Asset, which can be used
	// for unmarshalling an asset from its binary representation.
	NewAsset() Asset

	// NewAppID returns an object of type AppID, which can be used for
	// unmarshalling an app identifier from its binary representation.
	NewAppID() AppID
}

// SetBackend sets the global channel backend. Must not be called directly but
// through importing the needed backend.
func SetBackend(b Backend) {
	if backend != nil {
		panic("channel backend already set")
	}
	backend = b
}

// CalcID calculates the CalcID.
func CalcID(p *Params) ID {
	return backend.CalcID(p)
}

// Sign creates a signature from the account a on state s.
func Sign(a wallet.Account, s *State) (wallet.Sig, error) {
	return backend.Sign(a, s)
}

// Verify verifies that a signature was a valid signature from addr on a state.
func Verify(addr wallet.Address, state *State, sig wallet.Sig) (bool, error) {
	return backend.Verify(addr, state, sig)
}

// NewAsset returns a variable of type Asset, which can be used
// for unmarshalling an asset from its binary representation.
func NewAsset() Asset {
	return backend.NewAsset()
}

// NewAppID returns an object of type AppID, which can be used for
// unmarshalling an app identifier from its binary representation.
func NewAppID() AppID {
	return backend.NewAppID()
}
