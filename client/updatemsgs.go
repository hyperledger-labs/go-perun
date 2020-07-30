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
	"io"

	"perun.network/go-perun/channel"
	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

func init() {
	wire.RegisterDecoder(wire.ChannelUpdate,
		func(r io.Reader) (wire.Msg, error) {
			var m msgChannelUpdate
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.ChannelUpdateAcc,
		func(r io.Reader) (wire.Msg, error) {
			var m msgChannelUpdateAcc
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.ChannelUpdateRej,
		func(r io.Reader) (wire.Msg, error) {
			var m msgChannelUpdateRej
			return &m, m.Decode(r)
		})
}

type (
	// ChannelMsg are all messages that can be routed to a particular channel
	// controller.
	ChannelMsg interface {
		wire.Msg
		ID() channel.ID
	}

	channelUpdateResMsg interface {
		ChannelMsg
		Ver() uint64
	}

	// msgChannelUpdate is the wire message of a channel update proposal. It
	// additionally holds the signature on the proposed state.
	msgChannelUpdate struct {
		ChannelUpdate
		// Sig is the signature on the proposed state by the peer sending the
		// ChannelUpdate.
		Sig wallet.Sig
	}

	// msgChannelUpdateAcc is the wire message sent as a positive reply to a
	// ChannelUpdate.  It references the channel ID and version and contains the
	// signature on the accepted new state by the sender.
	msgChannelUpdateAcc struct {
		// ChannelID is the channel ID.
		ChannelID channel.ID
		// Version of the state that is accepted.
		Version uint64
		// Sig is the signature on the proposed new state by the sender.
		Sig wallet.Sig
	}

	// msgChannelUpdateRej is the wire message sent as a negative reply to a
	// ChannelUpdate.  It references the channel ID and version and states a
	// reason for the rejection.
	msgChannelUpdateRej struct {
		// ChannelID is the channel ID.
		ChannelID channel.ID
		// Version of the state that is accepted.
		Version uint64
		// Reason states why the sender rejectes the proposed new state.
		Reason string
	}
)

var (
	_ ChannelMsg          = (*msgChannelUpdate)(nil)
	_ channelUpdateResMsg = (*msgChannelUpdateAcc)(nil)
	_ channelUpdateResMsg = (*msgChannelUpdateRej)(nil)
)

// Type returns this message's type: ChannelUpdate.
func (*msgChannelUpdate) Type() wire.Type {
	return wire.ChannelUpdate
}

// Type returns this message's type: ChannelUpdateAcc.
func (*msgChannelUpdateAcc) Type() wire.Type {
	return wire.ChannelUpdateAcc
}

// Type returns this message's type: ChannelUpdateRej.
func (*msgChannelUpdateRej) Type() wire.Type {
	return wire.ChannelUpdateRej
}

func (c msgChannelUpdate) Encode(w io.Writer) error {
	return perunio.Encode(w, c.State, c.ActorIdx, c.Sig)
}

func (c *msgChannelUpdate) Decode(r io.Reader) (err error) {
	if c.State == nil {
		c.State = new(channel.State)
	}
	if err := perunio.Decode(r, c.State, &c.ActorIdx); err != nil {
		return err
	}
	c.Sig, err = wallet.DecodeSig(r)
	return err
}

func (c msgChannelUpdateAcc) Encode(w io.Writer) error {
	return perunio.Encode(w, c.ChannelID, c.Version, c.Sig)
}

func (c *msgChannelUpdateAcc) Decode(r io.Reader) (err error) {
	if err := perunio.Decode(r, &c.ChannelID, &c.Version); err != nil {
		return err
	}
	c.Sig, err = wallet.DecodeSig(r)
	return err
}

func (c msgChannelUpdateRej) Encode(w io.Writer) error {
	return perunio.Encode(w, c.ChannelID, c.Version, c.Reason)
}

func (c *msgChannelUpdateRej) Decode(r io.Reader) (err error) {
	return perunio.Decode(r, &c.ChannelID, &c.Version, &c.Reason)
}

// ID returns the id of the channel this update refers to.
func (c *msgChannelUpdate) ID() channel.ID {
	return c.State.ID
}

// ID returns the id of the channel this update acceptance refers to.
func (c *msgChannelUpdateAcc) ID() channel.ID {
	return c.ChannelID
}

// ID returns the id of the channel this update rejection refers to.
func (c *msgChannelUpdateRej) ID() channel.ID {
	return c.ChannelID
}

// Ver returns the version of the state this update acceptance refers to.
func (c *msgChannelUpdateAcc) Ver() uint64 {
	return c.Version
}

// Ver returns the version of the state this update rejection refers to.
func (c *msgChannelUpdateRej) Ver() uint64 {
	return c.Version
}
