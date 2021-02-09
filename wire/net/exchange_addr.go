// Copyright 2021 - See NOTICE file for copyright holders.
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

package net

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	pkg "perun.network/go-perun/pkg/context"
	"perun.network/go-perun/wire"
)

// AuthenticationError describes an error which occures when the ExchangeAddrs
// protcol fails because it got a different Address than expected.
type AuthenticationError struct {
	Sender, Receiver, Own wire.Address
}

// NewAuthenticationError creates a new AuthenticationError.
func NewAuthenticationError(sender, receiver, own wire.Address, msg string) error {
	return errors.Wrap(&AuthenticationError{
		Sender:   sender,
		Receiver: receiver,
		Own:      own,
	}, msg)
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("failed authentication (Sender: %v, Receiver: %v, Own: %v)", e.Sender, e.Receiver, e.Own)
}

// IsAuthenticationError returns true if the error was a AuthenticationError.
func IsAuthenticationError(err error) bool {
	cause := errors.Cause(err)
	_, ok := cause.(*AuthenticationError)
	return ok
}

// ExchangeAddrsActive executes the active role of the address exchange
// protocol. It is executed by the person that dials.
//
// In the future, it will be extended to become a proper authentication
// protocol. The protocol will then exchange Perun addresses and establish
// authenticity.
func ExchangeAddrsActive(ctx context.Context, id wire.Account, peer wire.Address, conn Conn) error {
	var err error
	ok := pkg.TerminatesCtx(ctx, func() {
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
			err = NewAuthenticationError(e.Sender, e.Recipient, id.Address(), "unmatched response sender or recipient")
		}
	})

	if !ok {
		// nolint:errcheck,gosec
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
	ok := pkg.TerminatesCtx(ctx, func() {
		var e *wire.Envelope
		if e, err = conn.Recv(); err != nil {
			err = errors.WithMessage(err, "receiving auth message")
		} else if _, ok := e.Msg.(*wire.AuthResponseMsg); !ok {
			err = errors.Errorf("expected AuthResponse wire msg, got %v", e.Msg.Type())
		} else if !e.Recipient.Equals(id.Address()) {
			err = NewAuthenticationError(e.Sender, e.Recipient, id.Address(), "unmatched response sender or recipient")
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
		// nolint:errcheck,gosec
		conn.Close()
		return nil, errors.WithMessage(ctx.Err(), "timeout")
	} else if err != nil {
		// nolint:errcheck,gosec
		conn.Close()
	}
	return addr, err
}
