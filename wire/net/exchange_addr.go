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
	"perun.network/go-perun/channel/persistence/test"

	"github.com/pkg/errors"

	"perun.network/go-perun/wire"
	pkg "polycry.pt/poly-go/context"
)

// AuthenticationError describes an error which occures when the ExchangeAddrs
// protcol fails because it got a different Address than expected.
type AuthenticationError struct {
	Sender, Receiver, Own map[int]wire.Address
}

// NewAuthenticationError creates a new AuthenticationError.
func NewAuthenticationError(sender, receiver, own map[int]wire.Address, msg string) error {
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
func ExchangeAddrsActive(ctx context.Context, id wire.Account, peer map[int]wire.Address, conn Conn) error {
	var err error
	ok := pkg.TerminatesCtx(ctx, func() {
		authMsg, err2 := wire.NewAuthResponseMsg(id)
		if err2 != nil {
			err = errors.WithMessage(err2, "creating auth message")
			return
		}
		err = conn.Send(&wire.Envelope{
			Sender:    id.Address(),
			Recipient: peer,
			Msg:       authMsg,
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
		} else if check := VerifyAddressSignature(peer, e.Msg.(*wire.AuthResponseMsg).Signature); check != nil {
			err = errors.WithMessage(err, "verifying peer address's signature")
		} else if !test.EqualWireMaps(e.Recipient, id.Address()) &&
			!test.EqualWireMaps(e.Sender, peer) {
			err = NewAuthenticationError(e.Sender, e.Recipient, id.Address(), "unmatched response sender or recipient")
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
func ExchangeAddrsPassive(ctx context.Context, id wire.Account, conn Conn) (map[int]wire.Address, error) {
	var addr map[int]wire.Address
	var err error
	ok := pkg.TerminatesCtx(ctx, func() {
		var e *wire.Envelope
		if e, err = conn.Recv(); err != nil {
			err = errors.WithMessage(err, "receiving auth message")
		} else if _, ok := e.Msg.(*wire.AuthResponseMsg); !ok {
			err = errors.Errorf("expected AuthResponse wire msg, got %v", e.Msg.Type())
		} else if !test.EqualWireMaps(e.Recipient, id.Address()) {
			err = NewAuthenticationError(e.Sender, e.Recipient, id.Address(), "unmatched response sender or recipient")
		} else if err = VerifyAddressSignature(e.Sender, e.Msg.(*wire.AuthResponseMsg).Signature); err != nil {
			err = errors.WithMessage(err, "verifying peer address's signature")
		}

		if err != nil {
			return
		}

		authMsg, err2 := wire.NewAuthResponseMsg(id)
		if err2 != nil {
			err = errors.WithMessage(err2, "creating auth message")
			return
		}
		addr, err = e.Sender, conn.Send(&wire.Envelope{
			Sender:    id.Address(),
			Recipient: e.Sender,
			Msg:       authMsg,
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

// VerifyAddressSignature verifies a signature against the hash of an address.
// It relies on the MarshalBinary method of the provided wire.Address interface to generate the address hash.
// In case the MarshalBinary method doesn't produce the expected hash, the verification may fail.
func VerifyAddressSignature(addrs map[int]wire.Address, sig []byte) error {
	var addressBytes []byte
	addressBytes = append(addressBytes, byte(len(addrs)))
	for _, addr := range addrs {
		addrBytes, err := addr.MarshalBinary()
		if err != nil {
			return fmt.Errorf("failed to marshal address: %w", err)
		}
		addressBytes = append(addressBytes, addrBytes...)
	}
	for _, addr := range addrs {
		err := addr.Verify(addressBytes, sig)
		if err != nil {
			return err
		}
	}
	return nil
}
