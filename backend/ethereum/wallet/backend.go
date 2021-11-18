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
	"io"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"perun.network/go-perun/wallet"
	perunio "polycry.pt/poly-go/io"
)

// Backend implements the utility interface defined in the wallet package.
type Backend struct{}

// SigLen length of a signature in byte.
// ref https://godoc.org/github.com/ethereum/go-ethereum/crypto/secp256k1#Sign
// ref https://github.com/ethereum/go-ethereum/blob/54b271a86dd748f3b0bcebeaf678dc34e0d6177a/crypto/signature_cgo.go#L66
const SigLen = 65

// sigVSubtract value that is subtracted from the last byte of a signature if
// the last bytes exceeds it.
const sigVSubtract = 27

// compile-time check that the ethereum backend implements the perun backend.
var _ wallet.Backend = (*Backend)(nil)

// DecodeAddress decodes an address from an io.Reader.
func (*Backend) DecodeAddress(r io.Reader) (wallet.Address, error) {
	return DecodeAddress(r)
}

// DecodeSig reads a []byte with length of an ethereum signature.
func (*Backend) DecodeSig(r io.Reader) (wallet.Sig, error) {
	return DecodeSig(r)
}

// VerifySignature verifies a signature.
func (*Backend) VerifySignature(msg []byte, sig wallet.Sig, a wallet.Address) (bool, error) {
	return VerifySignature(msg, sig, a)
}

// DecodeAddress decodes an address from an io.Reader.
func DecodeAddress(r io.Reader) (wallet.Address, error) {
	addr := new(Address)
	return addr, addr.Decode(r)
}

// DecodeSig reads a []byte with length of an ethereum signature.
func DecodeSig(r io.Reader) (wallet.Sig, error) {
	buf := make(wallet.Sig, SigLen)
	return buf, perunio.Decode(r, &buf)
}

// VerifySignature verifies if a signature was made by this account.
func VerifySignature(msg []byte, sig wallet.Sig, a wallet.Address) (bool, error) {
	hash := PrefixedHash(msg)
	sigCopy := make([]byte, SigLen)
	copy(sigCopy, sig)
	if len(sigCopy) == SigLen && (sigCopy[SigLen-1] >= sigVSubtract) {
		sigCopy[SigLen-1] -= sigVSubtract
	}
	pk, err := crypto.SigToPub(hash, sigCopy)
	if err != nil {
		return false, errors.WithStack(err)
	}
	addr := crypto.PubkeyToAddress(*pk)
	return a.Equals((*Address)(&addr)), nil
}

// PrefixedHash adds an ethereum specific prefix to the hash of given data, rehashes the results
// and returns it.
func PrefixedHash(data []byte) []byte {
	hash := crypto.Keccak256(data)
	prefix := []byte("\x19Ethereum Signed Message:\n32")
	return crypto.Keccak256(prefix, hash)
}
