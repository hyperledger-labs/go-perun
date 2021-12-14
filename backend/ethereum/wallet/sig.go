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

package wallet

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"perun.network/go-perun/wallet"
)

const (
	// sigLen length of a signature in byte.
	// ref https://godoc.org/github.com/ethereum/go-ethereum/crypto/secp256k1#Sign
	// ref https://github.com/ethereum/go-ethereum/blob/54b271a86dd748f3b0bcebeaf678dc34e0d6177a/crypto/signature_cgo.go#L66
	sigLen = 65

	// sigVSubtract value that is subtracted from the last byte of a signature if
	// the last bytes exceeds it.
	sigVSubtract = 27
)

// Sig represents a signature generated using an ethereum account.
type Sig []byte

// MarshalBinary marshals the signature into its binary representation. Error
// will always be nil, it is for implementing BinaryMarshaler.
func (s Sig) MarshalBinary() ([]byte, error) {
	return s[:], nil
}

// UnmarshalBinary unmarshals the signature from its binary representation.
func (s *Sig) UnmarshalBinary(data []byte) error {
	if len(data) != sigLen {
		return fmt.Errorf("unexpected signature length %d, want %d", len(data), sigLen) //nolint: goerr113
	}
	copy(*s, data)
	return nil
}

// Verify verifies if the signature on the given message was made by the given
// address.
func (s *Sig) Verify(msg []byte, addr wallet.Address) (bool, error) {
	hash := PrefixedHash(msg)
	sigCopy := make([]byte, sigLen)
	copy(sigCopy, *s)
	if sigCopy[sigLen-1] >= sigVSubtract {
		sigCopy[sigLen-1] -= sigVSubtract
	}

	pk, err := crypto.SigToPub(hash, sigCopy)
	if err != nil {
		return false, errors.WithStack(err)
	}
	addrInSig := crypto.PubkeyToAddress(*pk)
	return addr.Equal((*Address)(&addrInSig)), nil
}

// Clone returns a deep copy of the signature.
func (s Sig) Clone() wallet.Sig {
	clone := Sig(make([]byte, sigLen))
	copy(clone, s)
	return &clone
}
