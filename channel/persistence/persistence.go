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
		ChannelCreated(ctx context.Context, source Source, peers []peer.Address) error

		// ChannelRemoved is called by the client when a channel is removed because
		// it has been successfully settled and its data is no longer needed. All
		// data associated with this channel may be discarded.
		ChannelRemoved(ctx context.Context, id channel.ID) error

		// Staged is called when a new valid state got set as the new staging
		// state. It may already contain one valid signature, either by a remote
		// peer or us locally. Hence, this only needs to persist a channel's staged
		// state, all its currently known signatures and the phase.
		Staged(context.Context, Source) error

		// SigAdded is called when a new signature is added to the current staging
		// state. Only the signature for the given index needs to be persisted.
		SigAdded(context.Context, Source, channel.Index) error

		// Enabled is called when the current state is updated to the staging state.
		// The old current state may be discarded. The current state and phase
		// should be persisted.
		Enabled(context.Context, Source) error

		// PhaseChanged is called when a phase change occurred that did not change
		// the current or staging transaction. Only the phase needs to be persisted.
		PhaseChanged(context.Context, Source) error

		// Close is called by the client when it shuts down. No more persistence
		// requests will be made after this call and the Persister should free up
		// all possible resources.
		io.Closer
	}

	// Source is a source of channel data. It allows access to all information
	// needed for persistence. The ID, Idx and Params only need to be persisted
	// once per channel as they stay constant during a channel's lifetime.
	Source interface {
		ID() channel.ID                 // ID is the channel ID of this source. It is the same as Params().ID().
		Idx() channel.Index             // Idx is the own index in the channel.
		Params() *channel.Params        // Params are the channel parameters.
		StagingTX() channel.Transaction // StagingTX is the staged transaction (State+incomplete list of sigs).
		CurrentTX() channel.Transaction // CurrentTX is the current transaction (State+complete list of sigs).
		Phase() channel.Phase           // Phase is the phase in which the channel is currently in.
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
		// Peers are the channel network peers of this channel. They must have the
		// same ordering as the Params.Peers, omitting the own peer.
		Peers     []peer.Address
		Idx       channel.Index       // Idx is the own index in the channel.
		Params    *channel.Params     // Params are the channel parameters.
		StagingTX channel.Transaction // StagingTx is the staging transaction.
		CurrentTX channel.Transaction // CurrentTX is the current transaction.
		Phase     channel.Phase       // Phase is the current channel phase.
	}
)

// CloneSource creates a new Channel object whose fields are clones of the data
// coming from Source s.
func CloneSource(s Source) *Channel {
	return &Channel{
		Idx:       s.Idx(),
		Params:    s.Params().Clone(),
		StagingTX: s.StagingTX().Clone(),
		CurrentTX: s.CurrentTX().Clone(),
		Phase:     s.Phase(),
	}
}
