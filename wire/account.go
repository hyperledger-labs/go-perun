// Copyright 2025 - See NOTICE file for copyright holders.
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
	"encoding/binary"
	"fmt"
	"io"
	"math"

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
// authenticity within the Perun peer-to-peer network.
type Account interface {
	// Address used by this account.
	Address() Address

	// Sign signs the given message with this account's private key.
	Sign(msg []byte) ([]byte, error)
}

var _ Msg = (*AuthResponseMsg)(nil)

// AuthResponseMsg is the response message in the peer authentication protocol.
type AuthResponseMsg struct {
	Signature []byte
}

// Type returns AuthResponse.
func (m *AuthResponseMsg) Type() Type {
	return AuthResponse
}

// Encode encodes this AuthResponseMsg into an io.Writer.
// It writes the signature to the writer.
func (m *AuthResponseMsg) Encode(w io.Writer) error {
	// Write the length of the signature
	l := len(m.Signature)
	if l > math.MaxUint32 {
		return fmt.Errorf("signature length out of bounds: %d", len(m.Signature))
	}
	err := binary.Write(w, binary.BigEndian, uint32(l))
	if err != nil {
		return fmt.Errorf("failed to write signature length: %w", err)
	}
	// Write the signature itself
	_, err = w.Write(m.Signature)
	return err
}

// Decode decodes an AuthResponseMsg from an io.Reader.
// It reads the signature from the reader.
func (m *AuthResponseMsg) Decode(r io.Reader) (err error) {
	// Read the length of the signature
	var signatureLen uint32
	err = binary.Read(r, binary.BigEndian, &signatureLen)
	if err != nil {
		return fmt.Errorf("failed to read signature length: %w", err)
	}

	// Read the signature bytes
	m.Signature = make([]byte, signatureLen)
	_, err = io.ReadFull(r, m.Signature)
	if err != nil {
		return fmt.Errorf("failed to read signature: %w", err)
	}
	return nil
}

// NewAuthResponseMsg creates an authentication response message.
func NewAuthResponseMsg(acc map[wallet.BackendID]Account, backendID wallet.BackendID) (Msg, error) {
	addressMap := make(map[wallet.BackendID]Address)
	for id, a := range acc {
		addressMap[id] = a.Address()
	}
	var addressBytes []byte
	addressBytes = append(addressBytes, byte(len(addressMap)))
	for _, addr := range addressMap {
		addrBytes, err := addr.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal address: %w", err)
		}
		addressBytes = append(addressBytes, addrBytes...)
	}
	signature, err := acc[backendID].Sign(addressBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign address: %w", err)
	}

	return &AuthResponseMsg{
		Signature: signature,
	}, nil
}
