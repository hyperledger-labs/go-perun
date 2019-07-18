// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import (
	"io"
	"strconv"

	"github.com/pkg/errors"
)

// PeerMsg objects are messages that are sent between peers, but do not belong
// to a specific state channel, such as channel creation requests.
type PeerMsg interface {
	Msg
	// Type returns the message's implementing type.
	Type() PeerMsgType
}

func decodePeerMsg(reader io.Reader) (PeerMsg, error) {
	var Type PeerMsgType
	if err := Type.Decode(reader); err != nil {
		return nil, errors.Wrap(err, "failed to read the message type")
	}

	switch Type {
	default:
		panic("decodePeerMsg(): Unhandled peer message type: " + Type.String())
	}
}

func encodePeerMsg(msg PeerMsg, writer io.Writer) error {
	if err := msg.Type().Encode(writer); err != nil {
		return errors.Wrap(err, "failed to write the message type")
	}

	if err := msg.encode(writer); err != nil {
		return errors.Wrap(err, "failed to write the message contents")
	}

	return nil
}

// peerMsg allows default-implementing the Category function in peer messages.
type peerMsg struct {
	channelID ChannelID
}

func (*peerMsg) Category() Category {
	return Peer
}

// PeerMsgType is an enumeration used for (de)serializing channel messages
// and identifying a channel message's type.
//
// When changing this type, also change encode() and decode().
type PeerMsgType uint8

// Enumeration of channel message types.
const (
	// A dummy message, replace with real message types.
	DummyPeerMsg PeerMsgType = iota
	peerMsgTypeEnd
)

func (t PeerMsgType) String() string {
	if !t.Valid() {
		return strconv.Itoa(int(t))
	}
	return []string{
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
	buf := [1]byte{}
	if _, err := reader.Read(buf[:]); err != nil {
		return errors.Wrap(err, "failed to read channel message type")
	}
	*t = PeerMsgType(buf[0])
	if !t.Valid() {
		return errors.New("invalid channel message type encoding: " + t.String())
	}
	return nil
}
