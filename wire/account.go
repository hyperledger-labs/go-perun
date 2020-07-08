// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"context"
	"io"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wallet"
)

func init() {
	RegisterDecoder(AuthResponse,
		func(r io.Reader) (Msg, error) {
			var m AuthResponseMsg
			return &m, m.Decode(r)
		})
}

// Account is a node's permanent Perun identity, which is used to establish
// authenticity within the Perun peer-to-peer network. For now, it is just a
// stub.
type Account = wallet.Account

// ExchangeAddrsActive executes the active role of the address exchange
// protocol. It is executed by the person that dials.
//
// In the future, it will be extended to become a proper authentication
// protocol. The protocol will then exchange Perun addresses and establish
// authenticity.
func ExchangeAddrsActive(ctx context.Context, id Account, peer Address, conn Conn) error {
	var err error
	ok := test.TerminatesCtx(ctx, func() {
		err = conn.Send(&Envelope{
			Sender:    id.Address(),
			Recipient: peer,
			Msg:       NewAuthResponseMsg(id),
		})
		if err != nil {
			err = errors.WithMessage(err, "sending message")
			return
		}

		var e *Envelope
		if e, err = conn.Recv(); err != nil {
			err = errors.WithMessage(err, "receiving message")
		} else if _, ok := e.Msg.(*AuthResponseMsg); !ok {
			err = errors.Errorf("expected AuthResponse wire msg, got %v", e.Msg.Type())
		} else if !e.Recipient.Equals(id.Address()) &&
			!e.Sender.Equals(peer) {
			err = errors.Errorf("unmatched response sender or recipient")
		}
	})

	if !ok {
		conn.Close()
		return errors.WithMessage(ctx.Err(), "timeout")
	}

	return err
}

// ExchangeAddrsPassive executes the passive role of the address exchange
// protocol. It is executed by the person that listens for incoming connections.
func ExchangeAddrsPassive(ctx context.Context, id Account, conn Conn) (Address, error) {
	var addr Address
	var err error
	ok := test.TerminatesCtx(ctx, func() {
		var e *Envelope
		if e, err = conn.Recv(); err != nil {
			err = errors.WithMessage(err, "receiving auth message")
		} else if _, ok := e.Msg.(*AuthResponseMsg); !ok {
			err = errors.Errorf("expected AuthResponse wire msg, got %v", e.Msg.Type())
		} else if !e.Recipient.Equals(id.Address()) {
			err = errors.Errorf("unmatched response sender or recipient")
		}
		if err != nil {
			return
		}
		addr, err = e.Sender, conn.Send(&Envelope{
			Sender:    id.Address(),
			Recipient: e.Sender,
			Msg:       NewAuthResponseMsg(id),
		})
	})

	if !ok {
		conn.Close()
		return nil, errors.WithMessage(ctx.Err(), "timeout")
	} else if err != nil {
		conn.Close()
	}
	return addr, err
}

var _ Msg = (*AuthResponseMsg)(nil)

// AuthResponseMsg is the response message in the peer authentication protocol.
//
// This will be expanded later to contain signatures.
type AuthResponseMsg struct {
}

// Type returns AuthResponse.
func (m *AuthResponseMsg) Type() Type {
	return AuthResponse
}

// Encode encodes this AuthResponseMsg into an io.Writer.
func (m *AuthResponseMsg) Encode(w io.Writer) error {
	return nil
}

// Decode decodes an AuthResponseMsg from an io.Reader.
func (m *AuthResponseMsg) Decode(r io.Reader) (err error) {
	return nil
}

// NewAuthResponseMsg creates an authentication response message.
func NewAuthResponseMsg(_ Account) Msg {
	return &AuthResponseMsg{}
}
