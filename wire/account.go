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
	"crypto/sha256"
	"encoding/binary"
	"fmt"
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

	// Sign signs the given message with this account's private key.
	Sign(msg []byte) ([]byte, error)
}

var _ Msg = (*AuthResponseMsg)(nil)

// AuthResponseMsg is the response message in the peer authentication protocol.
//
// This will be expanded later to contain signatures.
type AuthResponseMsg struct {
	SignatureSize uint32 // Length of the signature
	Signature     []byte
}

// Type returns AuthResponse.
func (m *AuthResponseMsg) Type() Type {
	return AuthResponse
}

// Encode encodes this AuthResponseMsg into an io.Writer.
func (m *AuthResponseMsg) Encode(w io.Writer) error {
	// Write the signature size first
	if err := encodeUint32(w, m.SignatureSize); err != nil {
		return err
	}

	// Write the signature
	_, err := w.Write(m.Signature)
	return err
}

// Decode decodes an AuthResponseMsg from an io.Reader.
func (m *AuthResponseMsg) Decode(r io.Reader) error {
	var err error
	// Read the signature size first
	if m.SignatureSize, err = decodeUint32(r); err != nil {
		return err
	}

	// Read the signature
	m.Signature = make([]byte, m.SignatureSize)
	if _, err := io.ReadFull(r, m.Signature); err != nil {
		return fmt.Errorf("failed to read signature: %w", err)
	}
	return nil
}

// NewAuthResponseMsg creates an authentication response message.
func NewAuthResponseMsg(acc Account) (Msg, error) {
	addressBytes, err := acc.Address().MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal address: %w", err)
	}
	hashed := sha256.Sum256(addressBytes)
	signature, err := acc.Sign(hashed[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign address: %w", err)
	}

	return &AuthResponseMsg{
		SignatureSize: uint32(len(signature)),
		Signature:     signature,
	}, nil
}

// encodeUint32 encodes a uint32 value into an io.Writer.
func encodeUint32(w io.Writer, v uint32) error {
	sigSize := 4 // uint32 size
	buf := make([]byte, sigSize)
	binary.BigEndian.PutUint32(buf, v)
	_, err := w.Write(buf)
	return err
}

// decodeUint32 decodes a uint32 value from an io.Reader.
func decodeUint32(r io.Reader) (uint32, error) {
	sigSize := 4 // uint32 size
	buf := make([]byte, sigSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buf), nil
}
