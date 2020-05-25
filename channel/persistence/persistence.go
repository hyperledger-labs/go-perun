// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package persistence

import (
	"context"
	"io"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/peer"
)

type (
	// A Persister is used by the framework to persist channel data during
	// different steps of a channel's lifetime. It is guaranteed by the framework
	// that, per channel, only one of those methods is called concurrently.
	// However, for different channels, several calls may be made concurrently and
	// independently.
	Persister interface {
		// ChannelCreated is called by the client when a new channel is created,
		// before funding it. This should fully persist all of the source's data.
		// The current state will be the fully signed version 0 state. The staging
		// state will be empty. The passed peers are the channel network peers,
		// which should also be persisted.
		ChannelCreated(ctx context.Context, source channel.Source, peers []peer.Address) error

		// ChannelRemoved is called by the client when a channel is removed because
		// it has been successfully settled and its data is no longer needed. All
		// data associated with this channel may be discarded.
		ChannelRemoved(ctx context.Context, id channel.ID) error

		// Staged is called when a new valid state got set as the new staging
		// state. It may already contain one valid signature, either by a remote
		// peer or us locally. Hence, this only needs to persist a channel's staged
		// state, all its currently known signatures and the phase.
		Staged(context.Context, channel.Source) error

		// SigAdded is called when a new signature is added to the current staging
		// state. Only the signature for the given index needs to be persisted.
		SigAdded(context.Context, channel.Source, channel.Index) error

		// Enabled is called when the current state is updated to the staging state.
		// The old current state may be discarded. The current state and phase
		// should be persisted.
		Enabled(context.Context, channel.Source) error

		// PhaseChanged is called when a phase change occurred that did not change
		// the current or staging transaction. Only the phase needs to be persisted.
		PhaseChanged(context.Context, channel.Source) error

		// Close is called by the client when it shuts down. No more persistence
		// requests will be made after this call and the Persister should free up
		// all possible resources.
		io.Closer
	}

	// A Restorer allows a Client to restore channel machines. It has methods that
	// return iterators over channel data.
	Restorer interface {
		// RestoreAll should return an iterator over all persisted channels.
		RestoreAll() (ChannelIterator, error)

		// RestorePeer should return an iterator over all persisted channels which
		// the given peer is a part of.
		RestorePeer(peer.Address) (ChannelIterator, error)
	}

	// PersistRestorer is a Persister and Restorer on the same data source and
	// data sink.
	PersistRestorer interface {
		Persister
		Restorer
	}

	// A ChannelIterator is an iterator over Channels, i.e., channel data that is
	// necessary for restoring a channel machine. It needs to be implemented by a
	// persistence backend to allow the framework to restore channels.
	ChannelIterator interface {
		// Next should restore the next persisted channel. If no channel was found,
		// or the context times out, it should return false.
		Next(context.Context) bool

		// Channel should return the latest channel data that was restored via Next.
		// It is guaranteed by the framework to only be called after Next returned
		// true.
		Channel() *Channel

		// Close is called by the framework when Next returned false, or when
		// prematurely aborting. If an error occurred during the last Next call,
		// this error should be returned. Otherwise, nil should be returned if the
		// iterator was exhausted or closed prematurely by this Close call. This
		// call should free up all resources used by the iterator.
		io.Closer
	}

	// A Channel holds all data that is necessary for restoring a channel
	// controller.
	Channel struct {
		IdxV       channel.Index       // IdxV is the own index in the channel.
		ParamsV    *channel.Params     // ParamsV are the channel parameters.
		StagingTXV channel.Transaction // StagingTxV is the staging transaction.
		CurrentTXV channel.Transaction // CurrentTXV is the current transaction.
		PhaseV     channel.Phase       // PhaseV is the current channel phase.
	}
)

var _ channel.Source = (*Channel)(nil)

// CloneSource creates a new Channel object whose fields are clones of the data
// coming from Source s.
func CloneSource(s channel.Source) *Channel {
	return &Channel{
		IdxV:       s.Idx(),
		ParamsV:    s.Params().Clone(),
		StagingTXV: s.StagingTX().Clone(),
		CurrentTXV: s.CurrentTX().Clone(),
		PhaseV:     s.Phase(),
	}
}

// ID is the channel ID of this source. It is the same as Params().ID().
func (c *Channel) ID() channel.ID { return c.ParamsV.ID() }

// Idx is the own index in the channel.
func (c *Channel) Idx() channel.Index { return c.IdxV }

// Params are the channel parameters.
func (c *Channel) Params() *channel.Params { return c.ParamsV }

// StagingTX is the staged transaction (State+incomplete list of sigs).
func (c *Channel) StagingTX() channel.Transaction { return c.StagingTXV }

// CurrentTX is the current transaction (State+complete list of sigs).
func (c *Channel) CurrentTX() channel.Transaction { return c.CurrentTXV }

// Phase is the phase in which the channel is currently in.
func (c *Channel) Phase() channel.Phase { return c.PhaseV }
