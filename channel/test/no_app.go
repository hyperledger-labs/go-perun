// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/channel/test"

import (
	"io"
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

// NoApp Is a placeholder `ActionApp` and `StateApp` that does nothing
type NoApp struct {
	definition wallet.Address
}

// NoAppData the app data of the `NoApp`. It still needs a data field, otherwise the serializeable
// tests fail because it succeeds on a closed socket.
type NoAppData struct {
	Value int64
}

var _ channel.ActionApp = new(NoApp)
var _ channel.StateApp = new(NoApp)
var _ channel.Data = new(NoAppData)

func NewNoAppData(value int64) *NoAppData {
	return &NoAppData{Value: value}
}

func newRandomNoAppData(rng *rand.Rand) *NoAppData {
	return NewNoAppData(rng.Int63())
}

func (a NoAppData) Encode(w io.Writer) error {
	return wire.Encode(w, a.Value)
}

func (a *NoAppData) Decode(r io.Reader) error {
	return wire.Decode(r, &a.Value)
}

func (a NoAppData) Clone() channel.Data {
	return &NoAppData{Value: a.Value}
}

// NewNoApp return a new `NoApp`
func NewNoApp(definition wallet.Address) *NoApp {
	return &NoApp{definition: definition}
}

// Def returns nil
func (a NoApp) Def() wallet.Address {
	return a.definition
}

// ValidTransition returns nil
func (a NoApp) ValidTransition(*channel.Params, *channel.State, *channel.State) error {
	return nil
}

// ValidInit returns nil
func (a NoApp) ValidInit(*channel.Params, *channel.State) error {
	return nil
}

// ValidAction returns nil
func (a NoApp) ValidAction(*channel.Params, *channel.State, uint, channel.Action) error {
	return nil
}

// ApplyActions returns nil, nil
func (a NoApp) ApplyActions(params *channel.Params, state *channel.State, actions []channel.Action) (*channel.State, error) {
	newState := state.Clone()
	newState.Version++

	return newState, nil
}

// InitState returns channel.Allocation{}, nil, nil
func (a NoApp) InitState(*channel.Params, []channel.Action) (channel.Allocation, channel.Data, error) {
	return channel.Allocation{}, nil, nil
}

// DecodeData decodes a `NoAppData` from the reader and returns it
func (NoApp) DecodeData(r io.Reader) (channel.Data, error) {
	var data NoAppData
	return &data, data.Decode(r)
}

// DecodeAction return nil, nil
func (NoApp) DecodeAction(r io.Reader) (channel.Action, error) {
	return nil, nil
}
