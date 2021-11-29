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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"io"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
)

var curve = elliptic.P256()

// Backend implements the utility interface defined in the wallet package.
type Backend struct{}

var _ wallet.Backend = new(Backend)

// DecodeAddress decodes an address from the given Reader.
func (b *Backend) DecodeAddress(r io.Reader) (wallet.Address, error) {
	var addr Address
	return &addr, addr.Decode(r)
}

// DecodeSig reads a []byte with length of a signature.
func (b *Backend) DecodeSig(r io.Reader) (wallet.Sig, error) {
	buf := make(wallet.Sig, (curve.Params().BitSize/bitsPerByte)*pointsPerSig)
	return buf, perunio.Decode(r, &buf)
}

// VerifySignature verifies if a signature was made by this account.
func (b *Backend) VerifySignature(msg []byte, sig wallet.Sig, a wallet.Address) (bool, error) {
	addr, ok := a.(*Address)
	if !ok {
		log.Panic("Wrong address type passed to Backend.VerifySignature")
	}
	pk := (*ecdsa.PublicKey)(addr)

	r, s, err := deserializeSignature(sig)
	if err != nil {
		return false, errors.WithMessage(err, "could not deserialize signature")
	}

	// escda.Verify needs a digest as input
	// ref https://golang.org/pkg/crypto/ecdsa/#Verify
	return ecdsa.Verify(pk, digest(msg), r, s), nil
}

func digest(msg []byte) []byte {
	digest := sha256.Sum256(msg)
	return digest[:]
}
