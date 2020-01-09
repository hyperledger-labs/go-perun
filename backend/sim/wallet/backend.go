// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package wallet // import "perun.network/go-perun/backend/sim/wallet"

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"io"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

var curve = elliptic.P256()

// Backend implements the utility interface defined in the wallet package.
type Backend struct{}

var _ wallet.Backend = new(Backend)

// DecodeAddress decodes an address from the given Reader
func (b *Backend) DecodeAddress(r io.Reader) (wallet.Address, error) {
	var addr Address
	return &addr, addr.Decode(r)
}

// DecodeSig reads a []byte with length of a signature
func (b *Backend) DecodeSig(r io.Reader) (wallet.Sig, error) {
	buf := make(wallet.Sig, curve.Params().BitSize/4)
	return buf, wire.Decode(r, &buf)
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
