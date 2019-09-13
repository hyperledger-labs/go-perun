// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel // import "perun.network/go-perun/backend/sim/channel"

import (
	"io"

	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"

	"perun.network/go-perun/channel"
	perun "perun.network/go-perun/wallet"
)

// noApp Is a placeholder `ActionApp` and `StateApp` that does nothing
type noApp struct {
	definition wallet.Address
}

type dummyData struct {
	Flag bool
}

var _ channel.ActionApp = new(noApp)
var _ channel.StateApp = new(noApp)
var _ channel.Data = new(dummyData)

func newDummyData(flag bool) *dummyData {
	return &dummyData{Flag: flag}
}

func (a dummyData) Encode(w io.Writer) error {
	return wire.Encode(w, a.Flag)
}

func (a *dummyData) Decode(r io.Reader) error {
	return wire.Decode(r, &a.Flag)
}

func (a dummyData) Clone() channel.Data {
	return &dummyData{Flag: a.Flag}
}

// newNoApp return a new `NoApp`
func newNoApp(definition wallet.Address) *noApp {
	return &noApp{definition: definition}
}

// Def returns nil
func (a noApp) Def() perun.Address {
	return a.definition
}

// ValidTransition returns nil
func (a noApp) ValidTransition(*channel.Params, *channel.State, *channel.State) error {
	return nil
}

// ValidInit returns nil
func (a noApp) ValidInit(*channel.Params, *channel.State) error {
	return nil
}

// ValidAction returns nil
func (a noApp) ValidAction(*channel.Params, *channel.State, uint, channel.Action) error {
	return nil
}

// ApplyActions returns nil, nil
func (a noApp) ApplyActions(params *channel.Params, state *channel.State, actions []channel.Action) (*channel.State, error) {
	newState := state.Clone()
	newState.Version++

	return newState, nil
}

// InitState returns channel.Allocation{}, nil, nil
func (a noApp) InitState(*channel.Params, []channel.Action) (channel.Allocation, channel.Data, error) {
	return channel.Allocation{}, nil, nil
}
