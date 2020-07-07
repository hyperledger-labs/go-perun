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

// ExchangeAddrs exchanges Perun addresses of peers. It's the initial protocol
// that is run when a new peer connection is established. It returns the address
// of the peer on the other end of the connection. If the supplied context times
// out before the protocol finishes, closes the connection.
//
// In the future, ExchangeAddrs will be replaced by Authenticate to run a proper
// authentication protocol. The protocol will then exchange Perun addresses and
// establish authenticity.
func ExchangeAddrs(ctx context.Context, id Account, conn Conn) (Address, error) {
	var addr Address
	var err error
	ok := test.TerminatesCtx(ctx, func() {
		sent := make(chan error, 1)
		go func() { sent <- conn.Send(NewAuthResponseMsg(id)) }()

		var m Msg
		if m, err = conn.Recv(); err != nil {
			err = errors.WithMessage(err, "receiving message")
		} else if addrM, ok := m.(*AuthResponseMsg); !ok {
			err = errors.Errorf("expected AuthResponse wire msg, got %v", m.Type())
		} else {
			err = <-sent // Wait until the message was sent.
			addr = addrM.Address
		}
	})

	if !ok {
		conn.Close()
		return nil, ctx.Err()
	}

	return addr, err
}

var _ Msg = (*AuthResponseMsg)(nil)

// AuthResponseMsg is the response message in the peer authentication protocol.
type AuthResponseMsg struct {
	Address Address
}

// Type returns AuthResponse.
func (m *AuthResponseMsg) Type() Type {
	return AuthResponse
}

// Encode encodes this AuthResponseMsg into an io.Writer.
func (m *AuthResponseMsg) Encode(w io.Writer) error {
	return m.Address.Encode(w)
}

// Decode decodes an AuthResponseMsg from an io.Reader.
func (m *AuthResponseMsg) Decode(r io.Reader) (err error) {
	m.Address, err = wallet.DecodeAddress(r)
	return
}

// NewAuthResponseMsg creates an authentication response message.
// In the future, it will also take an authentication challenge message as
// additional argument.
func NewAuthResponseMsg(id Account) Msg {
	return &AuthResponseMsg{id.Address()}
}
