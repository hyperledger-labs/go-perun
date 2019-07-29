// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import (
	"strconv"
	"io"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
)

// PeerMsg objects are messages that are sent between peers, but do not belong
// to a specific state channel, such as channel creation requests.
type PeerMsg interface {
	Msg
	// Type returns the message's implementing type.
	Type() PeerMsgType
}

func decodePeerMsg(reader io.Reader) (msg PeerMsg, err error) {
	var Type PeerMsgType
	if err := Type.Decode(reader); err != nil {
		return nil, errors.WithMessage(err, "failed to read the message type")
	}

	switch Type {
	case PeerDummy:
		msg = &DummyPeerMsg{}
	default:
		log.Panicf("decodePeerMsg(): Unhandled peer message type: %v", Type)
	}

	if err := msg.decode(reader); err != nil {
		return nil, errors.WithMessagef(err, "failed to decode %v", Type)
	}
	return msg, nil
}

func encodePeerMsg(msg PeerMsg, writer io.Writer) error {
	if err := msg.Type().Encode(writer); err != nil {
		return errors.WithMessage(err, "failed to write the message type")
	}

	if err := msg.encode(writer); err != nil {
		return errors.WithMessage(err, "failed to write the message contents")
	}

	return nil
}

// peerMsg allows default-implementing the Category function in peer messages.
type peerMsg struct {
	// Currently empty, until we know what peer messages actually look like.
}

func (*peerMsg) Category() Category {
	return Peer
}

// PeerMsgType is an enumeration used for (de)serializing channel messages
// and identifying a channel message's type.
//
// When changing this type, also change Encode() and Decode().
type PeerMsgType uint8

// Enumeration of channel message types.
const (
	// This is a dummy peer message. It is used for testing purposes until the
	// first actual peer message type is added.
	PeerDummy PeerMsgType = iota

	// This constant marks the first invalid enum value.
	peerMsgTypeEnd
)

// String returns the name of a peer message type, if it is valid, otherwise,
// returns its numerical representation for debugging purposes.
func (t PeerMsgType) String() string {
	if !t.Valid() {
		return strconv.Itoa(int(t))
	}
	return [...]string{
		"DummyPeerMsg",
	}[t]
}

// Valid checks whether a PeerMsgType is a valid value.
func (t PeerMsgType) Valid() bool {
	return t < peerMsgTypeEnd
}

func (t PeerMsgType) Encode(writer io.Writer) error {
	if _, err := writer.Write([]byte{byte(t)}); err != nil {
		return errors.Wrap(err, "failed to write channel message type")
	}
	return nil
}

func (t *PeerMsgType) Decode(reader io.Reader) error {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return errors.WithMessage(err, "failed to read channel message type")
	}
	*t = PeerMsgType(buf[0])
	if !t.Valid() {
		return errors.New("invalid channel message type encoding: " + t.String())
	}
	return nil
}
