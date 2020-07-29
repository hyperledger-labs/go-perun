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

package client

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/wire"
)

var syncReplyTimeout = 10 * time.Second

func (c *Client) restorePeerChannels(ctx context.Context, p wire.Address) (err error) {
	it, err := c.pr.RestorePeer(p)
	if err != nil {
		return errors.WithMessagef(err, "restoring channels for peer: %v", err)
	}
	defer func() {
		if cerr := it.Close(); cerr != nil {
			err = errors.WithMessagef(err, "(error closing iterator: %v)", cerr)
		}
	}()

	// Serially restore channels. We might change this to parallel restoring once
	// we initiate the sync protocol from here again.
	for it.Next(ctx) {
		chdata := it.Channel()
		if err := c.restoreChannel(p, chdata); err != nil {
			return errors.WithMessage(err, "restoring channel")
		}
	}
	return nil
}

func (c *Client) restoreChannel(p wire.Address, chdata *persistence.Channel) error {
	log := c.logChan(chdata.ID())
	log.Debug("Restoring channel...")

	// TODO:
	// Send outgoing channel sync request and receive possibly newer channel data.
	// Incoming sync requests are handled by handleSyncMsg which is called from
	// the client's request loop.

	// TODO: read peers from chdata when available
	peers := make([]wire.Address, 2)
	peers[chdata.IdxV] = c.address
	peers[chdata.IdxV^1] = p

	// Create the channel's controller.
	ch, err := c.channelFromSource(chdata, peers...)
	if err != nil {
		return errors.WithMessage(err, "creating channel controller")
	}

	// Putting the channel into the channel registry will call the
	// OnNewChannel callback so that the user can deal with the restored
	// channel.
	if !c.channels.Put(chdata.ID(), ch) {
		log.Warn("Channel already present, closing restored channel.")
		// If the channel already existed, close this one.
		// nolint:errcheck,gosec
		ch.Close()
	}
	log.Info("Channel restored.")
	return nil
}

// handleSyncMsg is the passive incoming sync message handler. If the channel
// exists, it just sends the current channel data to the requester. If the
// own channel is in the Signing phase, the ongoing update is discarded so that
// the channel is reverted to the Acting phase.
func (c *Client) handleSyncMsg(peer wire.Address, msg *msgChannelSync) {
	log := c.logChan(msg.ID()).WithField("peer", peer)
	ch, ok := c.channels.Get(msg.ID())
	if !ok {
		log.Error("received sync message for unknown channel")
		return
	}

	// TODO: cancel ongoing protocol, like Update

	ctx, cancel := context.WithTimeout(c.Ctx(), syncReplyTimeout)
	defer cancel()
	// Lock machine while replying to sync request.
	if !ch.machMtx.TryLockCtx(ctx) {
		log.Errorf("Could not lock machine mutex in time: %v", ctx.Err())
	}
	defer ch.machMtx.Unlock()

	syncMsg := newChannelSyncMsg(persistence.CloneSource(ch.machine))
	if err := c.conn.pubMsg(ctx, syncMsg, peer); err != nil {
		log.Error("Error sending sync reply: ", err)
		return
	}
	cancel() // can already release context resourcers

	// Revert ongoing update since this is how synchronization is currently
	// implemented... the peer will do the same.
	if ch.machine.Phase() == channel.Signing {
		// The passed context is used for persistence, so use client life-time context
		if err := ch.machine.DiscardUpdate(c.Ctx()); err != nil {
			log.Error("Error discarding update: ", err)
		}
	}
}

// syncChannel synchronizes the channel state with the given peer and modifies
// the current state if required.
// nolint:unused
func (c *Client) syncChannel(ctx context.Context, ch *persistence.Channel, p wire.Address) (err error) {
	recv := wire.NewReceiver()
	// nolint:errcheck
	defer recv.Close() // ignore error
	id := ch.ID()
	err = c.conn.Subscribe(recv, func(m *wire.Envelope) bool {
		return m.Msg.Type() == wire.ChannelSync && m.Msg.(ChannelMsg).ID() == id
	})
	if err != nil {
		return errors.WithMessage(err, "subscribing on relay")
	}

	sendError := make(chan error, 1)
	// syncMsg needs to be a clone so that there's no data race when updating the
	// own channel data later.
	syncMsg := newChannelSyncMsg(persistence.CloneSource(ch))
	go func() { sendError <- c.conn.pubMsg(ctx, syncMsg, p) }()
	defer func() {
		// When returning, either log the send error, or return it.
		sendErr := <-sendError
		if err == nil {
			err = errors.WithMessage(sendErr, "sending sync message")
		} else if err != nil && sendErr != nil {
			c.logChan(id).Errorf("Error sending sync message: %v", sendErr)
		}
	}()

	// Receive sync message.
	env, err := recv.Next(ctx)
	if err != nil {
		return errors.WithMessage(err, "receiving sync message")
	}
	msg := env.Msg.(*msgChannelSync) // safe by the predicate
	// Validate sync message.
	if err := validateMessage(ch, msg); err != nil {
		return errors.WithMessage(err, "invalid message")
	}
	// Merge restored state with received state.
	if msg.CurrentTX.Version > ch.CurrentTXV.Version {
		ch.CurrentTXV = msg.CurrentTX
	}

	return revisePhase(ch)
}

// validateMessage validates the remote channel sync message.
// nolint:unused, nestif
func validateMessage(ch *persistence.Channel, msg *msgChannelSync) error {
	v := ch.CurrentTX().Version
	mv := msg.CurrentTX.Version

	if msg.CurrentTX.ID != ch.ID() {
		return errors.New("channel ID mismatch")
	}
	if mv == v {
		if err := msg.CurrentTX.State.Equal(ch.CurrentTX().State); err != nil {
			return errors.WithMessage(err, "different states for same version")
		}
	} else if mv > v {
		// Validate the received message first.
		if len(msg.CurrentTX.Sigs) != len(ch.Params().Parts) {
			return errors.New("sigs length mismatch")
		}
		for i, sig := range msg.CurrentTX.Sigs {
			ok, err := channel.Verify(ch.Params().Parts[i], ch.Params(), msg.CurrentTX.State, sig)
			if err != nil {
				return errors.WithMessagef(err, "validating sig %d", i)
			}
			if !ok {
				return errors.Errorf("invalid sig %d", i)
			}
		}
	}
	return nil
}

// nolint:unused
func revisePhase(ch *persistence.Channel) error {
	// nolint: gocritic
	if ch.PhaseV <= channel.Funding && ch.CurrentTXV.Version == 0 {
		return errors.New("channel in Funding phase - funding during restore not implemented yet")
		// if version > 0, phase will be set to Acting/Final at the end
	} else if ch.PhaseV > channel.Final && ch.PhaseV < channel.Withdrawn {
		// looks like an abort settlement
		return errors.New("settling channel restored")
	} else if ch.PhaseV == channel.Withdrawn {
		// This channel is already settled
		return nil
	}

	// Reset potential Signing phase
	if ch.CurrentTXV.IsFinal {
		ch.PhaseV = channel.Final
	}
	ch.PhaseV = channel.Acting
	return nil
}
