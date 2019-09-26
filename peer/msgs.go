// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"encoding/binary"
	"io"
	"math"

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

type Commitment = []byte

type Proposal struct {
	ChallengeDuration uint64
	Commit            Commitment
	EphemeralAddr     wallet.Address
	AppDef            wallet.Address
	InitData          channel.Data
	InitBals          channel.Allocation
}

func (p Proposal) Category() wiremsg.Category {
	return wiremsg.Peer
}

func (p Proposal) encode(w io.Writer) error {
	if err := wire.Encode(w, p.ChallengeDuration); err != nil {
		return errors.WithMessagef(err, "Challenge duration encoding")
	}

	if len(p.Commit) > math.MaxInt32 {
		return errors.Errorf(
			"Expected maximum commitment length of %d bytes, got %d",
			math.MaxInt32, len(p.Commit))
	}

	commitLen := int32(len(p.Commit))
	if err := wire.Encode(w, commitLen); err != nil {
		return errors.WithMessagef(err, "Proposal commitment length encoding")
	}
	if _, err := w.Write(p.Commit); err != nil {
		return errors.WithMessagef(err, "Proposal commitment encoding")
	}

	if err := perunIo.Encode(w, p.EphemeralAddr, p.AppDef, p.InitData, &p.InitBals); err != nil {
		return errors.WithMessagef(err, "Proposal encoding")
	}

	return nil
}

func (p *Proposal) decode(r io.Reader) error {
	if err := wire.Decode(r, &p.ChallengeDuration); err != nil {
		return errors.WithMessagef(err, "Challenge duration decoding")
	}

	var commitLen int32
	if err := binary.Read(r, binary.LittleEndian, &commitLen); err != nil {
		return errors.WithMessagef(err, "Commitment length decoding")
	} else if commitLen < 0 {
		return errors.WithMessagef(err, "Negative decoded commitment length")
	}

	p.Commit = make([]byte, commitLen)

	if n, err := io.ReadFull(r, p.Commit); n < int(commitLen) || err != nil {
		return errors.WithMessagef(
			err, "Expected reading %d bytes, got %d", commitLen, n)
	}

	if ephemeralAddr, err := wallet.DecodeAddress(r); err != nil {
		return errors.WithMessagef(err, "Ephemeral address decoding")
	} else {
		p.EphemeralAddr = ephemeralAddr
	}

	if appDef, err := wallet.DecodeAddress(r); err != nil {
		return errors.WithMessagef(err, "App address decoding")
	} else {
		p.AppDef = appDef
	}

	p.InitData = &channel.DummyData{}
	p.InitBals = channel.Allocation{}
	if err := perunIo.Decode(r, p.InitData, &p.InitBals); err != nil {
		return errors.WithMessagef(err, "Initial state decoding")
	}

	return nil
}

func (p *Proposal) Type() MsgType {
	return PeerProposal
}

type SessionID = [32]byte

type Response struct {
	SesID         SessionID
	Commit        Commitment
	EphemeralAddr wallet.Address
}

func (*Response) Category() wiremsg.Category {
	return wiremsg.Peer
}

func (*Response) Type() MsgType {
	return PeerResponse
}

func (r *Response) encode(w io.Writer) error {
	if _, err := w.Write(r.SesID[:]); err != nil {
		return errors.WithMessagef(err, "Response SID encoding")
	}

	if len(r.Commit) > math.MaxInt32 {
		return errors.Errorf(
			"Expected maximum response commitment length %d, got %d",
			math.MaxInt32, len(r.Commit))
	}

	commitLen := int32(len(r.Commit))
	if err := wire.Encode(w, commitLen); err != nil {
		return errors.WithMessagef(err, "Response commitment length encoding")
	}
	if _, err := w.Write(r.Commit); err != nil {
		return errors.WithMessagef(err, "Response commitment encoding")
	}

	if err := r.EphemeralAddr.Encode(w); err != nil {
		return errors.WithMessagef(err, "Response ephemeral address encoding")
	}

	return nil
}

func (response *Response) decode(r io.Reader) error {
	response.SesID = SessionID{}
	if _, err := io.ReadFull(r, response.SesID[:]); err != nil {
		return errors.WithMessagef(err, "Response SID decoding")
	}

	var commitLen int32
	if err := binary.Read(r, binary.LittleEndian, &commitLen); err != nil {
		return errors.WithMessagef(err, "Response commitment length decoding")
	}
	if commitLen < 0 {
		return errors.Errorf(
			"Decoded negative response commitment length %d", commitLen)
	}

	response.Commit = make([]byte, commitLen)
	if _, err := io.ReadFull(r, response.Commit); err != nil {
		return errors.WithMessagef(err, "Response commitment decoding")
	}

	if ephemeralAddr, err := wallet.DecodeAddress(r); err != nil {
		return errors.WithMessagef(err, "App address decoding")
	} else {
		response.EphemeralAddr = ephemeralAddr
	}

	return nil
}
