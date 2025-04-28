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

package libp2p

import (
	"crypto/sha256"
	"fmt"
	"math/rand"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"perun.network/go-perun/wire"
)

// Address is a peer address for wire discovery.
type Address struct {
	peer.ID
}

// NewAddress returns a new address.
func NewAddress(id peer.ID) *Address {
	return &Address{ID: id}
}

// Equal returns whether the two addresses are equal.
func (a *Address) Equal(b wire.Address) bool {
	bTyped, ok := b.(*Address)
	if !ok {
		panic("wrong type")
	}

	return a.ID == bTyped.ID
}

// Cmp compares the byte representation of two addresses. For `a.Cmp(b)`
// returns -1 if a < b, 0 if a == b, 1 if a > b.
func (a *Address) Cmp(b wire.Address) int {
	bTyped, ok := b.(*Address)
	if !ok {
		panic("wrong type")
	}
	if a.ID < bTyped.ID {
		return -1
	} else if a.ID == bTyped.ID {
		return 0
	}
	return 1

}

// NewRandomAddress returns a new random peer address.
func NewRandomAddress(rng *rand.Rand) *Address {
	_, publicKey, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rng)
	if err != nil {
		panic(err)
	}

	id, err := peer.IDFromPublicKey(publicKey)
	if err != nil {
		panic(err)
	}
	return &Address{id}
}

// Verify verifies the signature of a message.
func (a *Address) Verify(msg []byte, sig []byte) error {
	publicKey, err := a.ExtractPublicKey()
	if err != nil {
		return fmt.Errorf("extracting public key: %w", err)
	}

	hashed := sha256.Sum256(msg)

	b, err := publicKey.Verify(hashed[:], sig)
	if err != nil {
		return fmt.Errorf("verifying signature: %w", err)
	}
	if b {
		return nil
	}
	return fmt.Errorf("signature verification failed")
}
