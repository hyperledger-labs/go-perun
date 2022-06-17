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

package wire

import (
	"io"
)

func init() {
	RegisterDecoder(AuthResponse,
		func(r io.Reader) (Msg, error) {
			var m AuthResponseMsg
			return &m, m.Decode(r)
		})
}

// Account is a node's permanent Perun identity, which is used to establish
// authenticity within the Perun peer-to-peer network.
type Account interface {
	// Address used by this account.
	Address() Address
}

var _ Msg = (*AuthResponseMsg)(nil)

// AuthResponseMsg is the response message in the peer authentication protocol.
//
// This will be expanded later to contain signatures.
type AuthResponseMsg struct{}

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
