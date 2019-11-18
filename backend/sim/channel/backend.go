// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel // import "perun.network/go-perun/backend/sim/channel"

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"io"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

// backend implements the utility interface defined in the channel package.
type backend struct{}

var _ channel.Backend = new(backend)

// ChannelID calculates a channel's ID by hashing all fields of its parameters
func (*backend) ChannelID(p *channel.Params) channel.ID {
	w := sha256.New()

	// Write ChallengeDuration
	if err := wire.Encode(w, p.ChallengeDuration); err != nil {
		log.Panic("Could not serialize to buffer")
	}
	// Write Parts
	for _, addr := range p.Parts {
		if err := addr.Encode(w); err != nil {
			log.Panic("Could not write to sha256 hasher")
		}
	}
	// Write App Address
	if err := p.App.Def().Encode(w); err != nil {
		log.Panic("Could not write to sha256 hasher")
	}
	// Write Nonce
	if err := wire.Encode(w, p.Nonce); err != nil {
		log.Panic("Could not write to sha256 hasher")
	}

	var id channel.ID
	hash := w.Sum(nil)

	if copy(id[:], hash) != 32 {
		log.Panic("Could not copy id")
	}

	return id
}

// Sign signs `state`
func (b *backend) Sign(addr wallet.Account, params *channel.Params, state *channel.State) ([]byte, error) {
	if addr == nil || params == nil || state == nil {
		return nil, errors.New("argument nil")
	}
	log.Tracef("Signing state %s version %d", string(state.ID[:]), state.Version)

	buff := new(bytes.Buffer)
	w := bufio.NewWriter(buff)

	if err := b.encodeState(*state, w); err != nil {
		return nil, errors.WithMessage(err, "pack state")
	}

	if err := w.Flush(); err != nil {
		log.Panic("bufio flush")
	}

	return addr.SignData(buff.Bytes())
}

// Verify verifies the signature for `state`
func (b *backend) Verify(addr wallet.Address, params *channel.Params, state *channel.State, sig []byte) (bool, error) {
	if addr == nil || params == nil || state == nil {
		return false, errors.New("argument nil")
	}
	log.Tracef("Verifying state %s version %d", string(state.ID[:]), state.Version)

	buff := new(bytes.Buffer)
	w := bufio.NewWriter(buff)

	if err := b.encodeState(*state, w); err != nil {
		return false, errors.WithMessage(err, "pack state")
	}

	if err := w.Flush(); err != nil {
		log.Panic("bufio flush")
	}

	return wallet.VerifySignature(buff.Bytes(), sig, addr)
}

// encodeState packs all fields of a State into a []byte
func (b *backend) encodeState(s channel.State, w io.Writer) error {
	// Write ID
	if err := wire.ByteSlice(s.ID[:]).Encode(w); err != nil {
		return errors.WithMessage(err, "state id encode")
	}
	// Write Version
	if err := wire.Encode(w, s.Version); err != nil {
		return errors.WithMessage(err, "state version encode")
	}
	// Don't write the App Definition, since we do not want to sign it.
	// (The contract does not get the AppDef in the state and needs to verify the signature of it.)
	// Write Allocation
	if err := b.encodeAllocation(w, s.Allocation); err != nil {
		return errors.WithMessage(err, "state allocation encode")
	}
	// Write Data
	if err := s.Data.Encode(w); err != nil {
		return errors.WithMessage(err, "state data encode")
	}
	// Write IsFinal
	if err := wire.Encode(w, s.IsFinal); err != nil {
		return errors.WithMessage(err, "state isfinal encode")
	}

	return nil
}

// encodeAllocation Writes all fields of `a` to `w`
func (b *backend) encodeAllocation(w io.Writer, a channel.Allocation) error {
	// Write Assets
	for _, asset := range a.Assets {
		if err := asset.Encode(w); err != nil {
			return errors.WithMessage(err, "asset.Encode")
		}
	}
	// Write OfParts
	for _, ofpart := range a.OfParts {
		if err := b.encodeBals(w, ofpart); err != nil {
			return errors.WithMessage(err, "bals encode")
		}
	}
	// Write Locked
	for _, locked := range a.Locked {
		if err := b.encodeSubAlloc(w, locked); err != nil {
			return errors.WithMessage(err, "Alloc.Encode")
		}
	}

	return nil
}

// encodeSubAlloc Writes all fields of `s` to `w`
func (b *backend) encodeSubAlloc(w io.Writer, s channel.SubAlloc) error {
	// Write ID
	if err := wire.ByteSlice(s.ID[:]).Encode(w); err != nil {
		return errors.WithMessage(err, "ID encode")
	}
	// Write Bals
	if err := b.encodeBals(w, s.Bals); err != nil {
		return errors.WithMessage(err, "bals encode")
	}

	return nil
}

func (*backend) encodeBals(w io.Writer, bals []channel.Bal) error {
	for _, bal := range bals {
		if err := wire.Encode(w, bal); err != nil {
			return errors.WithMessage(err, "bal encode")
		}
	}

	return nil
}
