// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

// Package test provides a Persister implementation for testing purposes.
package test // import "perun.network/go-perun/channel/persistence/test"

import (
	"bytes"
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/peer"
)

// A Persister is a persistence.Persister implementation for testing purposes.
// It is create by passing a *testing.T to NewPersister. Besides the methods
// implementing Persister, it provides methods for asserting the currently
// persisted state of channels.
type Persister struct {
	t *testing.T

	chans map[channel.ID]*persistence.Channel
}

// NewPersister creates a new testing persister that reports assert errors on
// the passed *testing.T t.
func NewPersister(t *testing.T) *Persister {
	return &Persister{
		t:     t,
		chans: make(map[channel.ID]*persistence.Channel),
	}
}

// ChannelCreated is called by the client when a new channel is created,
// before funding it. This should fully persist all of the source's data.
// The current state will be the fully signed version 0 state. The staging
// state will be empty. The passed peers are the channel network peers,
// which should also be persisted.
func (p *Persister) ChannelCreated(
	_ context.Context, source channel.Source, peers []peer.Address) error {
	id := source.ID()
	_, ok := p.chans[id]
	if ok {
		return errors.Errorf("channel already persisted: %x", source.ID())
	}

	p.chans[id] = persistence.CloneSource(source)
	p.chans[id].Peers = peers // Peers are constant so we don't clone.
	return nil
}

// ChannelRemoved is called by the client when a channel is removed because
// it has been successfully settled and its data is no longer needed.
func (p *Persister) ChannelRemoved(_ context.Context, id channel.ID) error {
	_, ok := p.chans[id]
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", id)
	}
	delete(p.chans, id)
	return nil
}

// Staged is called when a new valid state got set as the new staging
// state. It may already contain one valid signature, either by a remote
// peer or us locally. Hence, this only needs to persist a channel's staged
// state, all its currently known signatures and the phase.
func (p *Persister) Staged(_ context.Context, s channel.Source) error {
	ch, ok := p.chans[s.ID()]
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", s.ID())
	}

	ch.StagingTX = s.StagingTX().Clone()
	ch.Phase = s.Phase()
	return nil
}

// SigAdded is called when a new signature is added to the current staging
// state. Only the signature for the given index needs to be persisted.
func (p *Persister) SigAdded(_ context.Context, s channel.Source, idx channel.Index) error {
	ch, ok := p.chans[s.ID()]
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", s.ID())
	} else if ch.StagingTX.State == nil {
		return errors.Errorf("no staging transaction set")
	}

	ch.StagingTX.Sigs[idx] = bytes.Repeat(s.StagingTX().Sigs[idx], 1)
	return nil
}

// Enabled is called when the current state is updated to the staging state.
// The old current state may be discarded.
func (p *Persister) Enabled(_ context.Context, s channel.Source) error {
	ch, ok := p.chans[s.ID()]
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", s.ID())
	}

	ch.StagingTX = s.StagingTX().Clone()
	ch.CurrentTX = s.CurrentTX().Clone()
	ch.Phase = s.Phase()
	return nil
}

// PhaseChanged is called when a phase change occurred that did not change
// the current or staging transaction. Only the phase needs to be persisted.
func (p *Persister) PhaseChanged(_ context.Context, s channel.Source) error {
	ch, ok := p.chans[s.ID()]
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", s.ID())
	}

	ch.Phase = s.Phase()
	return nil
}

// Close resets the persister, i.e., all intrenally persisted channel data is
// deleted. It can be reused afterwards.
func (p *Persister) Close() error {
	p.chans = make(map[channel.ID]*persistence.Channel)
	return nil
}

// AssertEqual asserts that a channel of the same ID got persisted and that all
// its data fields match the data coming from Source s.
func (p *Persister) AssertEqual(s channel.Source) {
	ch, ok := p.chans[s.ID()]
	if !ok {
		p.t.Errorf("channel doesn't exist: %x", s.ID())
		return
	}

	assert := assert.New(p.t)
	assert.Equal(s.Idx(), ch.Idx, "Idx mismatch")
	assert.Equal(s.Params(), ch.Params, "Params mismatch")
	assert.Equal(s.StagingTX(), ch.StagingTX, "StagingTX mismatch")
	assert.Equal(s.CurrentTX(), ch.CurrentTX, "CurrentTX mismatch")
	assert.Equal(s.Phase(), ch.Phase, "Phase mismatch")
}
