// Copyright 2025 - See NOTICE file for copyright holders.
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
	"errors"

	"perun.network/go-perun/wallet"
)

// backend is set to the global channel backend. Must not be set directly but
// through importing the needed backend.
var backend map[wallet.BackendID]Backend

// Backend is an interface that needs to be implemented for every blockchain.
// It provides basic functionalities to the framework.
type Backend interface {
	// CalcID infers the channel id of a channel from its parameters. Usually,
	// this should be a hash digest of some or all fields of the parameters.
	// In order to guarantee non-malleability of States, any parameters omitted
	// from the CalcID digest need to be signed together with the State in
	// Sign().
	CalcID(*Params) (ID, error)

	// Sign signs a channel's State with the given Account.
	// Returns the signature or an error.
	Sign(wallet.Account, *State) (wallet.Sig, error)

	// Verify verifies that the provided signature on the state belongs to the
	// provided address.
	Verify(addr wallet.Address, state *State, sig wallet.Sig) (bool, error)

	// NewAsset returns a variable of type Asset, which can be used
	// for unmarshalling an asset from its binary representation.
	NewAsset() Asset

	// NewAppID returns an object of type AppID, which can be used for
	// unmarshalling an app identifier from its binary representation.
	NewAppID() (AppID, error)
}

// SetBackend sets the global channel backend. Must not be called directly but
// through importing the needed backend.
func SetBackend(b Backend, id int) {
	if backend == nil {
		backend = make(map[wallet.BackendID]Backend)
	}
	if backend[wallet.BackendID(id)] != nil {
		panic("channel backend already set")
	}
	backend[wallet.BackendID(id)] = b
}

// CalcID calculates the CalcID.
func CalcID(p *Params) (map[wallet.BackendID]ID, error) {
	id := make(map[wallet.BackendID]ID)
	var err error
	for i := range p.Parts[0] {
		id[i], err = backend[i].CalcID(p)
		if err != nil {
			return nil, err
		}
	}
	return id, nil
}

// Sign creates a signature from the account a on state s.
func Sign(a wallet.Account, s *State, bID wallet.BackendID) (wallet.Sig, error) {
	return backend[bID].Sign(a, s)
}

// Verify verifies that a signature was a valid signature from addr on a state.
func Verify(a wallet.Address, state *State, sig wallet.Sig) (bool, error) {
	return backend[a.BackendID()].Verify(a, state, sig)
}

// NewAsset returns a variable of type Asset, which can be used
// for unmarshalling an asset from its binary representation.
func NewAsset(id wallet.BackendID) Asset {
	return backend[id].NewAsset()
}

// NewAppID returns an object of type AppID, which can be used for
// unmarshalling an app identifier from its binary representation.
func NewAppID() (AppID, error) {
	for i := range backend {
		id, err := backend[i].NewAppID()
		if err == nil {
			return id, nil
		}
	}
	return nil, errors.New("no backend found")
}
