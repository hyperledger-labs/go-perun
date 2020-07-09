// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"io"

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
