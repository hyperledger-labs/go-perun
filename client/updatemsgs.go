// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client

import (
	"io"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/msg"
)

func init() {
	msg.RegisterDecoder(msg.ChannelUpdate,
		func(r io.Reader) (msg.Msg, error) {
			var m ChannelUpdate
			return &m, m.Decode(r)
		})
	msg.RegisterDecoder(msg.ChannelUpdateAcc,
		func(r io.Reader) (msg.Msg, error) {
			var m ChannelUpdateAcc
			return &m, m.Decode(r)
		})
	msg.RegisterDecoder(msg.ChannelUpdateRej,
		func(r io.Reader) (msg.Msg, error) {
			var m ChannelUpdateRej
			return &m, m.Decode(r)
		})
}

type (
	// ChannelMsg are all messages that can be routed to a particular channel
	// controller.
	ChannelMsg interface {
		msg.Msg
		ID() channel.ID
	}

	// ChannelUpdate is a channel update proposal.
	ChannelUpdate struct {
		// State is the proposed new state.
		State *channel.State
		// ActorIdx is the actor causing the new state.  It does not need to
		// coincide with the sender of the request.
		ActorIdx uint16
		// Sig is the signature on the proposed state by the peer sending the
		// ChannelUpdate.
		Sig wallet.Sig
	}

	// ChannelUpdateAcc is sent as a positive reply to a ChannelUpdate.  It
	// references the channel ID and version and contains the signature on the
	// accepted new state by the sender.
	ChannelUpdateAcc struct {
		// ChannelID is the channel ID.
		ChannelID channel.ID
		// Version of the state that is accepted.
		Version uint64
		// Sig is the signature on the proposed new state by the sender.
		Sig wallet.Sig
	}

	// ChannelUpdateRej is sent as a negative reply to a ChannelUpdate.  It
	// references the channel ID and version and contains the signature on the
	// accepted new state by the sender.
	ChannelUpdateRej struct {
		// Reason states why the sender rejectes the proposed new state.
		Reason string
		// Alt is the proposed new alternative state with same version number as the
		// proposed state.
		Alt *channel.State
		// ActorIdx is the actor causing the new alternative state.  It does not
		// need to coincide with the sender of the rejection.
		ActorIdx uint16
		// Sig is the signature on the alternative state by the sender.
		Sig wallet.Sig
	}
)

var (
	_ ChannelMsg = (*ChannelUpdate)(nil)
	_ ChannelMsg = (*ChannelUpdateAcc)(nil)
	_ ChannelMsg = (*ChannelUpdateRej)(nil)
)

// Type returns this message's type: ChannelUpdate
func (*ChannelUpdate) Type() msg.Type {
	return msg.ChannelUpdate
}

// Type returns this message's type: ChannelUpdateAcc
func (*ChannelUpdateAcc) Type() msg.Type {
	return msg.ChannelUpdateAcc
}

// Type returns this message's type: ChannelUpdateRej
func (*ChannelUpdateRej) Type() msg.Type {
	return msg.ChannelUpdateRej
}

func (c ChannelUpdate) Encode(w io.Writer) error {
	return wire.Encode(w, c.State, c.ActorIdx, c.Sig)
}

func (c *ChannelUpdate) Decode(r io.Reader) (err error) {
	if c.State == nil {
		c.State = new(channel.State)
	}
	if err := wire.Decode(r, c.State, &c.ActorIdx); err != nil {
		return err
	}
	c.Sig, err = wallet.DecodeSig(r)
	return err
}

func (c ChannelUpdateAcc) Encode(w io.Writer) error {
	return wire.Encode(w, c.ChannelID, c.Version, c.Sig)
}

func (c *ChannelUpdateAcc) Decode(r io.Reader) (err error) {
	if err := wire.Decode(r, &c.ChannelID, &c.Version); err != nil {
		return err
	}
	c.Sig, err = wallet.DecodeSig(r)
	return err
}

func (c ChannelUpdateRej) Encode(w io.Writer) error {
	return wire.Encode(w, c.Reason, c.Alt, c.ActorIdx, c.Sig)
}

func (c *ChannelUpdateRej) Decode(r io.Reader) (err error) {
	if c.Alt == nil {
		c.Alt = new(channel.State)
	}
	if err := wire.Decode(r, &c.Reason, c.Alt, &c.ActorIdx); err != nil {
		return err
	}
	c.Sig, err = wallet.DecodeSig(r)
	return err
}

// ID returns the id of the channel this update refers to.
func (c *ChannelUpdate) ID() channel.ID {
	return c.State.ID
}

// ID returns the id of the channel this update acceptance refers to.
func (c *ChannelUpdateAcc) ID() channel.ID {
	return c.ChannelID
}

// ID returns the id of the channel this update rejection refers to.
func (c *ChannelUpdateRej) ID() channel.ID {
	return c.Alt.ID
}
