// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package client

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/wire"
)

func (c *Client) restorePeerChannels(p *peer.Peer, done func()) {
	log := c.logPeer(p)
	it, err := c.pr.RestorePeer(p.PerunAddress)
	if err != nil {
		log.Errorf("Failed to restore channels for peer: %v", err)
		p.Close()
	}

	var wg sync.WaitGroup
	wg.Add(1)
	for it.Next(c.Ctx()) {
		chdata := it.Channel()
		wg.Add(1)
		go func() {
			defer wg.Done()
			log := c.logChan(chdata.ID())
			log.Debug("Restoring channel...")
			// Synchronize the channel with the peer, and settle if this fails.
			if err := c.syncChannel(c.Ctx(), chdata, p); err != nil {
				log.Errorf("Error synchronizing channels: %v; attempting settlement...", err)
				chdata.PhaseV = channel.Withdrawing
				// No peers, because we don't want any connections.
				if ch, err := c.channelFromSource(chdata); err != nil {
					log.Errorf("Failed to reconstruct channel for settling: %v", err)
				} else if err := ch.Settle(c.Ctx()); err != nil {
					log.Errorf("Failed to settle channel: %v", err)
				}
				return
			}

			// Create the channel's controller.
			ch, err := c.channelFromSource(chdata, p)
			if err != nil {
				log.Errorf("Failed to restore channel: %v", err)
				return
			}
			// Putting the channel into the channel registry will call the
			// OnNewChannel callback so that the user can deal with the restored
			// channel.
			if !c.channels.Put(chdata.ID(), ch) {
				log.Warn("Channel already present, closing restored channel.")
				// If the channel already existed, close this one.
				ch.Close()
			} else {
				log.Info("Channel restored.")
			}
		}()
	}

	wg.Done()
	go func() { wg.Wait(); done() }()

	if err := it.Close(); err != nil {
		log.Errorf("Error while restoring a channel: %v", err)
	}
}

// syncChannel synchronizes the channel state with the given peer and modifies
// the current state if required.
func (c *Client) syncChannel(ctx context.Context, ch *persistence.Channel, p *peer.Peer) (err error) {
	recv := peer.NewReceiver()
	id := ch.ID()
	p.Subscribe(recv, func(m wire.Msg) bool {
		return m.Type() == wire.ChannelSync && m.(ChannelMsg).ID() == id
	})

	sendError := make(chan error, 1)
	// syncMsg needs to be a clone so that there's no data race when updating the
	// own channel data later.
	syncMsg := newChannelSyncMsg(persistence.CloneSource(ch))
	go func() { sendError <- p.Send(ctx, syncMsg) }()
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
	var _msg wire.Msg
	if _, _msg = recv.Next(ctx); _msg == nil {
		return errors.New("receiving sync message failed")
	}
	msg := _msg.(*msgChannelSync)
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
			ok, err := channel.Verify(
				ch.Params().Parts[i],
				ch.Params(),
				msg.CurrentTX.State,
				sig)
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

func revisePhase(ch *persistence.Channel) error {
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
