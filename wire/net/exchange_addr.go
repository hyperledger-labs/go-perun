// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package net

import (
	"context"

	"github.com/pkg/errors"
	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wire"
)

// ExchangeAddrsActive executes the active role of the address exchange
// protocol. It is executed by the person that dials.
//
// In the future, it will be extended to become a proper authentication
// protocol. The protocol will then exchange Perun addresses and establish
// authenticity.
func ExchangeAddrsActive(ctx context.Context, id wire.Account, peer wire.Address, conn Conn) error {
	var err error
	ok := test.TerminatesCtx(ctx, func() {
		err = conn.Send(&wire.Envelope{
			Sender:    id.Address(),
			Recipient: peer,
			Msg:       wire.NewAuthResponseMsg(id),
		})
		if err != nil {
			err = errors.WithMessage(err, "sending message")
			return
		}

		var e *wire.Envelope
		if e, err = conn.Recv(); err != nil {
			err = errors.WithMessage(err, "receiving message")
		} else if _, ok := e.Msg.(*wire.AuthResponseMsg); !ok {
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
func ExchangeAddrsPassive(ctx context.Context, id wire.Account, conn Conn) (wire.Address, error) {
	var addr wire.Address
	var err error
	ok := test.TerminatesCtx(ctx, func() {
		var e *wire.Envelope
		if e, err = conn.Recv(); err != nil {
			err = errors.WithMessage(err, "receiving auth message")
		} else if _, ok := e.Msg.(*wire.AuthResponseMsg); !ok {
			err = errors.Errorf("expected AuthResponse wire msg, got %v", e.Msg.Type())
		} else if !e.Recipient.Equals(id.Address()) {
			err = errors.Errorf("unmatched response sender or recipient")
		}
		if err != nil {
			return
		}
		addr, err = e.Sender, conn.Send(&wire.Envelope{
			Sender:    id.Address(),
			Recipient: e.Sender,
			Msg:       wire.NewAuthResponseMsg(id),
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
