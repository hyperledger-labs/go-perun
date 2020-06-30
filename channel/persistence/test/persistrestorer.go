// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

// Package test provides a PersistRestorer implementation for testing purposes
// as well as a generic PersistRestorer implementation test.
package test // import "perun.network/go-perun/channel/persistence/test"

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/peer"
)

// A PersistRestorer is a persistence.PersistRestorer implementation for testing purposes.
// It is create by passing a *testing.T to NewPersistRestorer. Besides the methods
// implementing PersistRestorer, it provides methods for asserting the currently
// persisted state of channels.
type PersistRestorer struct {
	t *testing.T

	mu    sync.RWMutex // protects chans map and peerChans access
	chans map[channel.ID]*persistence.Channel
	pcs   peerChans
}

// NewPersistRestorer creates a new testing PersistRestorer that reports assert
// errors on the passed *testing.T t.
func NewPersistRestorer(t *testing.T) *PersistRestorer {
	return &PersistRestorer{
		t:     t,
		chans: make(map[channel.ID]*persistence.Channel),
		pcs:   make(peerChans),
	}
}

// Persister implementation

// ChannelCreated fully persists all of the source's data.
func (p *PersistRestorer) ChannelCreated(
	_ context.Context, source channel.Source, peers []peer.Address) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := source.ID()
	_, ok := p.chans[id]
	if ok {
		return errors.Errorf("channel already persisted: %x", id)
	}

	p.chans[id] = persistence.CloneSource(source)
	p.pcs.Add(id, peers...)
	return nil
}

// ChannelRemoved removes the channel from the test persister's memory.
func (p *PersistRestorer) ChannelRemoved(_ context.Context, id channel.ID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, ok := p.chans[id]
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", id)
	}
	delete(p.chans, id)
	p.pcs.Delete(id)
	return nil
}

// Staged only persists a channel's staged state, all its currently known
// signatures and the phase.
func (p *PersistRestorer) Staged(_ context.Context, s channel.Source) error {
	ch, ok := p.get(s.ID())
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", s.ID())
	}

	ch.StagingTXV = s.StagingTX().Clone()
	ch.PhaseV = s.Phase()
	return nil
}

// SigAdded only persists the signature for the given index.
func (p *PersistRestorer) SigAdded(_ context.Context, s channel.Source, idx channel.Index) error {
	ch, ok := p.get(s.ID())
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", s.ID())
	} else if ch.StagingTXV.State == nil {
		return errors.Errorf("no staging transaction set")
	}

	ch.StagingTXV.Sigs[idx] = bytes.Repeat(s.StagingTX().Sigs[idx], 1)
	return nil
}

// Enabled fully persists the current and staging transaction and phase. The
// staging transaction should be nil.
func (p *PersistRestorer) Enabled(_ context.Context, s channel.Source) error {
	ch, ok := p.get(s.ID())
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", s.ID())
	}

	ch.StagingTXV = s.StagingTX().Clone()
	ch.CurrentTXV = s.CurrentTX().Clone()
	ch.PhaseV = s.Phase()
	return nil
}

// PhaseChanged only persists the phase.
func (p *PersistRestorer) PhaseChanged(_ context.Context, s channel.Source) error {
	ch, ok := p.get(s.ID())
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", s.ID())
	}

	ch.PhaseV = s.Phase()
	return nil
}

// Close resets the persister's memory, i.e., all internally persisted channel
// data is deleted. It can be reused afterwards.
func (p *PersistRestorer) Close() error {
	p.chans = make(map[channel.ID]*persistence.Channel)
	return nil
}

// AssertEqual asserts that a channel of the same ID got persisted and that all
// its data fields match the data coming from Source s.
func (p *PersistRestorer) AssertEqual(s channel.Source) {
	ch, ok := p.get(s.ID())
	if !ok {
		p.t.Errorf("channel doesn't exist: %x", s.ID())
		return
	}

	assert := assert.New(p.t)
	assert.Equal(s.Idx(), ch.IdxV, "Idx mismatch")
	assert.Equal(s.Params(), ch.ParamsV, "Params mismatch")
	assert.Equal(s.StagingTX(), ch.StagingTXV, "StagingTX mismatch")
	assert.Equal(s.CurrentTX(), ch.CurrentTXV, "CurrentTX mismatch")
	assert.Equal(s.Phase(), ch.PhaseV, "Phase mismatch")
}

// get is a mutexed access to the Channel stored at the given id.
// Since persister access is guaranteed to be single-threaded per channel, it
// makes sense for the Persister implementation methods to use this getter to
// get the pointer to the channel storage.
func (p *PersistRestorer) get(id channel.ID) (*persistence.Channel, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	ch, ok := p.chans[id]
	return ch, ok
}

// Restorer implementation

// ActivePeers returns all peers that channels are persisted for.
func (p *PersistRestorer) ActivePeers(context.Context) ([]peer.Address, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.pcs.Peers(), nil
}

// RestorePeer returns an iterator over all persisted channels which
// the given peer is a part of.
func (p *PersistRestorer) RestorePeer(peer peer.Address) (persistence.ChannelIterator, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ids := p.pcs.Get(peer)
	it := &chanIter{
		chans: make([]*persistence.Channel, len(ids)),
		idx:   -1,
	}
	for i, id := range ids {
		it.chans[i] = p.chans[id]
	}
	return it, nil
}

// RestoreChannel should return the channel with the requested ID.
func (p *PersistRestorer) RestoreChannel(_ context.Context, id channel.ID) (*persistence.Channel, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ch, ok := p.chans[id]
	if !ok {
		return nil, errors.Errorf("channel not found: %x", id)
	}

	return ch, nil
}

type chanIter struct {
	chans []*persistence.Channel
	idx   int // has to be initialized to -1
}

func (i *chanIter) Next(context.Context) bool {
	i.idx++
	return i.idx < len(i.chans)
}

func (i *chanIter) Channel() *persistence.Channel {
	return i.chans[i.idx]
}

func (i *chanIter) Close() error {
	i.chans = nil // GC
	return nil
}
