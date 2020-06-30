// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package persistence

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/peer"
)

// NonPersistRestorer is a PersistRestorer that doesn't do anything. All
// Persistence methods return nil and all Restorer methods return an empty
// iterator.
var NonPersistRestorer PersistRestorer = nonPersistRestorer{}

type nonPersistRestorer struct{}

// Persister implementation

func (nonPersistRestorer) ChannelCreated(context.Context, channel.Source, []peer.Address) error {
	return nil
}
func (nonPersistRestorer) ChannelRemoved(context.Context, channel.ID) error              { return nil }
func (nonPersistRestorer) Staged(context.Context, channel.Source) error                  { return nil }
func (nonPersistRestorer) SigAdded(context.Context, channel.Source, channel.Index) error { return nil }
func (nonPersistRestorer) Enabled(context.Context, channel.Source) error                 { return nil }
func (nonPersistRestorer) PhaseChanged(context.Context, channel.Source) error            { return nil }
func (nonPersistRestorer) Close() error                                                  { return nil }

// Restorer implementation

func (nonPersistRestorer) ActivePeers(context.Context) ([]peer.Address, error) { return nil, nil }

func (nonPersistRestorer) RestoreAll() (ChannelIterator, error) {
	return emptyChanIterator{}, nil
}

func (nonPersistRestorer) RestorePeer(peer.Address) (ChannelIterator, error) {
	return emptyChanIterator{}, nil
}
func (nonPersistRestorer) RestoreChannel(context.Context, channel.ID) (*Channel, error) {
	return nil, errors.New("channel not found")
}

type emptyChanIterator struct{}

func (emptyChanIterator) Next(context.Context) bool { return false }
func (emptyChanIterator) Channel() *Channel         { return nil }
func (emptyChanIterator) Close() error              { return nil }
