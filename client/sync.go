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
	"log"
	"time"

	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/wire"
)

var syncReplyTimeout = 10 * time.Second

// handleSyncMsg is the passive incoming sync message handler. If the channel
// exists, it just sends the current channel data to the requester. If the
// own channel is in the Signing phase, the ongoing update is discarded so that
// the channel is reverted to the Acting phase.
func (c *Client) handleSyncMsg(peer map[wallet.BackendID]wire.Address, msg *ChannelSyncMsg) {
	log := c.logChan(msg.ID()).WithField("peer", peer)
	ch, ok := c.channels.Channel(msg.ID())
	if !ok {
		log.Error("received sync message for unknown channel")
		return
	}

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
//
//nolint:unused
func (c *Client) syncChannel(ctx context.Context, ch *persistence.Channel, p map[wallet.BackendID]wire.Address) (err error) {
	recv := wire.NewReceiver()
	defer recv.Close() // ignore error
	id := ch.ID()
	err = c.conn.Subscribe(recv, func(m *wire.Envelope) bool {
		msg, ok := m.Msg.(*ChannelSyncMsg)
		return ok && channel.EqualIDs(msg.ID(), id)
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
	msg, ok := env.Msg.(*ChannelSyncMsg)
	if !ok {
		log.Panic("internal error: wrong message type")
	}
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
//
//nolint:unused, nestif
func validateMessage(ch *persistence.Channel, msg *ChannelSyncMsg) error {
	v := ch.CurrentTX().Version
	mv := msg.CurrentTX.Version

	if channel.EqualIDs(msg.CurrentTX.ID, ch.ID()) {
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
			for _, p := range ch.Params().Parts[i] {
				ok, err := channel.Verify(p, msg.CurrentTX.State, sig)
				if err != nil {
					return errors.WithMessagef(err, "validating sig %d", i)
				}
				if !ok {
					return errors.Errorf("invalid sig %d", i)
				}
			}
		}
	}
	return nil
}

//nolint:unused
func revisePhase(ch *persistence.Channel) error {
	//nolint:gocritic
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
