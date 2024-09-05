// Copyright 2020 - See NOTICE file for copyright holders.
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

// Package test provides a PersistRestorer implementation for testing purposes
// as well as a generic PersistRestorer implementation test.
package test // import "perun.network/go-perun/channel/persistence/test"

import (
	"bytes"
	"context"
	"perun.network/go-perun/wallet"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/wire"
)

// A PersistRestorer is a persistence.PersistRestorer implementation for testing purposes.
// It is create by passing a *testing.T to NewPersistRestorer. Besides the methods
// implementing PersistRestorer, it provides methods for asserting the currently
// persisted state of channels.
type PersistRestorer struct {
	t *testing.T

	mu    sync.RWMutex // protects chans map and peerChans access
	chans map[string]*persistence.Channel
	pcs   peerChans
}

// NewPersistRestorer creates a new testing PersistRestorer that reports assert
// errors on the passed *testing.T t.
func NewPersistRestorer(t *testing.T) *PersistRestorer {
	t.Helper()
	return &PersistRestorer{
		t:     t,
		chans: make(map[string]*persistence.Channel),
		pcs:   make(peerChans),
	}
}

// Persister implementation

// ChannelCreated fully persists all of the source's data.
func (pr *PersistRestorer) ChannelCreated(
	_ context.Context, source channel.Source, peers []map[wallet.BackendID]wire.Address, parent *map[wallet.BackendID]channel.ID,
) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	id := source.ID()
	_, ok := pr.chans[channel.IDKey(id)]
	if ok {
		return errors.Errorf("channel already persisted: %x", id)
	}

	pr.chans[channel.IDKey(id)] = persistence.FromSource(source, peers, parent)
	pr.pcs.Add(id, peers...)
	return nil
}

// ChannelRemoved removes the channel from the test persister's memory.
func (pr *PersistRestorer) ChannelRemoved(_ context.Context, id map[wallet.BackendID]channel.ID) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	_, ok := pr.chans[channel.IDKey(id)]
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", id)
	}
	delete(pr.chans, channel.IDKey(id))
	pr.pcs.Delete(id)
	return nil
}

// Staged only persists a channel's staged state, all its currently known
// signatures and the phase.
func (pr *PersistRestorer) Staged(_ context.Context, s channel.Source) error {
	ch, ok := pr.channel(s.ID())
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", s.ID())
	}

	ch.StagingTXV = s.StagingTX().Clone()
	ch.PhaseV = s.Phase()
	return nil
}

// SigAdded only persists the signature for the given index.
func (pr *PersistRestorer) SigAdded(_ context.Context, s channel.Source, idx channel.Index) error {
	ch, ok := pr.channel(s.ID())
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
func (pr *PersistRestorer) Enabled(_ context.Context, s channel.Source) error {
	ch, ok := pr.channel(s.ID())
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", s.ID())
	}

	ch.StagingTXV = s.StagingTX().Clone()
	ch.CurrentTXV = s.CurrentTX().Clone()
	ch.PhaseV = s.Phase()
	return nil
}

// PhaseChanged only persists the phase.
func (pr *PersistRestorer) PhaseChanged(_ context.Context, s channel.Source) error {
	ch, ok := pr.channel(s.ID())
	if !ok {
		return errors.Errorf("channel doesn't exist: %x", s.ID())
	}

	ch.PhaseV = s.Phase()
	return nil
}

// Close resets the persister's memory, i.e., all internally persisted channel
// data is deleted. It can be reused afterwards.
func (pr *PersistRestorer) Close() error {
	pr.chans = make(map[string]*persistence.Channel)
	return nil
}

// AssertEqual asserts that a channel of the same ID got persisted and that all
// its data fields match the data coming from Source s.
func (pr *PersistRestorer) AssertEqual(s channel.Source) {
	ch, ok := pr.channel(s.ID())
	if !ok {
		pr.t.Errorf("channel doesn't exist: %x", s.ID())
		return
	}

	assert := assert.New(pr.t)
	assert.Equal(s.Idx(), ch.IdxV, "Idx mismatch")
	assert.Equal(s.Params(), ch.ParamsV, "Params mismatch")
	assert.Equal(s.StagingTX(), ch.StagingTXV, "StagingTX mismatch")
	assert.Equal(s.CurrentTX(), ch.CurrentTXV, "CurrentTX mismatch")
	assert.Equal(s.Phase(), ch.PhaseV, "Phase mismatch")
}

// AssertNotExists asserts that a channel with the given ID does not exist.
func (pr *PersistRestorer) AssertNotExists(id map[wallet.BackendID]channel.ID) {
	_, ok := pr.channel(id)
	assert.Falsef(pr.t, ok, "channel shouldn't exist: %x", id)
}

// channel is a mutexed access to the Channel stored at the given id.
// Since persister access is guaranteed to be single-threaded per channel, it
// makes sense for the Persister implementation methods to use this getter to
// channel the pointer to the channel storage.
func (pr *PersistRestorer) channel(id map[wallet.BackendID]channel.ID) (*persistence.Channel, bool) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	ch, ok := pr.chans[channel.IDKey(id)]
	return ch, ok
}

// Restorer implementation

// ActivePeers returns all peers that channels are persisted for.
func (pr *PersistRestorer) ActivePeers(context.Context) ([]map[wallet.BackendID]wire.Address, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	return pr.pcs.Peers(), nil
}

// RestorePeer returns an iterator over all persisted channels which
// the given peer is a part of.
func (pr *PersistRestorer) RestorePeer(peer map[wallet.BackendID]wire.Address) (persistence.ChannelIterator, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	ids := pr.pcs.ID(peer)
	it := &chanIter{
		chans: make([]*persistence.Channel, len(ids)),
		idx:   -1,
	}
	for i, id := range ids {
		it.chans[i] = pr.chans[channel.IDKey(id)]
	}
	return it, nil
}

// RestoreChannel should return the channel with the requested ID.
func (pr *PersistRestorer) RestoreChannel(_ context.Context, id map[wallet.BackendID]channel.ID) (*persistence.Channel, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	ch, ok := pr.chans[channel.IDKey(id)]
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
