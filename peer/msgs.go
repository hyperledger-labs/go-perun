// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"encoding/binary"
	"io"
	"math"
	"math/big"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	perunIo "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
	wire "perun.network/go-perun/wire"
	wiremsg "perun.network/go-perun/wire/msg"
)

// DummyPeerMsg is a dummy message type used for testing.
type DummyPeerMsg struct {
	msg
	dummy int64
}

func (DummyPeerMsg) Type() MsgType {
	return PeerDummy
}

func (DummyPeerMsg) Category() wiremsg.Category {
	return wiremsg.Peer
}

func (m DummyPeerMsg) encode(writer io.Writer) error {
	return wire.Encode(writer, m.dummy)
}

func (m *DummyPeerMsg) decode(reader io.Reader) error {
	return wire.Decode(reader, &m.dummy)
}

type Nonce *big.Int

type Proposal struct {
	ChallengeDuration uint64
	Nonce             Nonce
	ParticipantAddr   wallet.Address
	AppDef            wallet.Address
	InitData          channel.Data
	InitBals          channel.Allocation
	Parts             []wallet.Address
}

func (p Proposal) Category() wiremsg.Category {
	return wiremsg.Peer
}

func (p Proposal) encode(w io.Writer) error {
	var nonceInt *big.Int
	nonceInt = p.Nonce
	if err := wire.Encode(w, p.ChallengeDuration, nonceInt); err != nil {
		return err
	}

	if err := perunIo.Encode(w, p.ParticipantAddr, p.AppDef, p.InitData, &p.InitBals); err != nil {
		return err
	}

	if len(p.Parts) > math.MaxInt32 {
		return errors.Errorf(
			"Expected maximum number of participants %d, got %d",
			math.MaxInt32, len(p.Parts))
	}

	numParts := int32(len(p.Parts))
	if err := binary.Write(w, binary.LittleEndian, numParts); err != nil {
		return err
	}
	ss := make([]perunIo.Serializable, len(p.Parts))
	for i := 0; i < len(p.Parts); i++ {
		ss[i] = p.Parts[i]
	}
	if err := perunIo.Encode(w, ss...); err != nil {
		return err
	}

	return nil
}

func (p *Proposal) decode(r io.Reader) error {
	var nonceInt *big.Int
	if err := wire.Decode(r, &p.ChallengeDuration, &nonceInt); err != nil {
		return err
	}
	p.Nonce = new(big.Int).Set(nonceInt)

	// read p.ParticipantAddr, p.AppDef
	if ephemeralAddr, err := wallet.DecodeAddress(r); err != nil {
		return err
	} else {
		p.ParticipantAddr = ephemeralAddr
	}
	if appDef, err := wallet.DecodeAddress(r); err != nil {
		return err
	} else {
		p.AppDef = appDef
	}

	p.InitData = &channel.DummyData{}
	p.InitBals = channel.Allocation{}

	if err := perunIo.Decode(r, p.InitData, &p.InitBals); err != nil {
		return err
	}

	var numParts int32
	if err := wire.Decode(r, &numParts); err != nil {
		return err
	}
	if numParts < 2 {
		return errors.Errorf(
			"Expected at least 2 participants, got %d", numParts)
	}

	p.Parts = make([]wallet.Address, numParts)
	for i := 0; i < len(p.Parts); i++ {
		if addr, err := wallet.DecodeAddress(r); err != nil {
			return err
		} else {
			p.Parts[i] = addr
		}
	}

	return nil
}

func (p *Proposal) Type() MsgType {
	return PeerProposal
}

type SessionID = [32]byte

type Response struct {
	SessID          SessionID
	ParticipantAddr wallet.Address
}

func (*Response) Category() wiremsg.Category {
	return wiremsg.Peer
}

func (*Response) Type() MsgType {
	return PeerResponse
}

func (r *Response) encode(w io.Writer) error {
	if _, err := w.Write(r.SessID[:]); err != nil {
		return errors.WithMessagef(err, "Response SID encoding")
	}

	if err := r.ParticipantAddr.Encode(w); err != nil {
		return errors.WithMessagef(err, "Response ephemeral address encoding")
	}

	return nil
}

func (response *Response) decode(r io.Reader) error {
	response.SessID = SessionID{}
	if _, err := io.ReadFull(r, response.SessID[:]); err != nil {
		return errors.WithMessagef(err, "Response SID decoding")
	}

	if ephemeralAddr, err := wallet.DecodeAddress(r); err != nil {
		return errors.WithMessagef(err, "App address decoding")
	} else {
		response.ParticipantAddr = ephemeralAddr
	}

	return nil
}
