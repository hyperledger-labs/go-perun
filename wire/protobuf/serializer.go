// Copyright 2022 - See NOTICE file for copyright holders.
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

package protobuf

import (
	"encoding/binary"
	"io"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wire"
)

// Serializer implements methods for encoding/decoding
// envelopes using protobuf encoding.
type Serializer struct{}

// func init() {
// 	wire.SetEnvelopeSerializer(Serializer{})
// }

// Encode encodes the envelope into the wire using perunio
// encoding format.
func (Serializer) Encode(w io.Writer, env *wire.Envelope) error { // nolint: funlen, cyclop, gocognit
	sender, err := env.Sender.MarshalBinary()
	if err != nil {
		return errors.WithMessage(err, "marshalling sender address")
	}
	recipient, err := env.Recipient.MarshalBinary()
	if err != nil {
		return errors.WithMessage(err, "marshalling recipient address")
	}

	var grpcMsg isEnvelope_Msg
	switch env.Msg.Type() {
	case wire.Ping:
		msg, ok := env.Msg.(*wire.PingMsg)
		if !ok {
			return errors.New("message type and content mismatch")
		}
		grpcMsg = &Envelope_PingMsg{
			PingMsg: &PingMsg{
				Created: msg.Created.UnixNano(),
			},
		}
	case wire.Pong:
		msg, ok := env.Msg.(*wire.PongMsg)
		if !ok {
			return errors.New("message type and content mismatch")
		}
		grpcMsg = &Envelope_PongMsg{
			PongMsg: &PongMsg{
				Created: msg.Created.UnixNano(),
			},
		}
	case wire.Shutdown:
		msg, ok := env.Msg.(*wire.ShutdownMsg)
		if !ok {
			return errors.New("message type and content mismatch")
		}
		grpcMsg = &Envelope_ShutdownMsg{
			ShutdownMsg: &ShutdownMsg{
				Reason: msg.Reason,
			},
		}
	case wire.AuthResponse:
		grpcMsg = &Envelope_AuthResponseMsg{
			AuthResponseMsg: &AuthResponseMsg{},
		}
	case wire.LedgerChannelProposal:
		msg, ok := env.Msg.(*client.LedgerChannelProposalMsg)
		if !ok {
			return errors.New("message type and content mismatch")
		}
		ledgerChannelProposal, err := fromLedgerChannelProposal(msg)
		if err != nil {
			return err
		}
		grpcMsg = &Envelope_LedgerChannelProposalMsg{
			LedgerChannelProposalMsg: ledgerChannelProposal,
		}
	case wire.SubChannelProposal:
		msg, ok := env.Msg.(*client.SubChannelProposalMsg)
		if !ok {
			return errors.New("message type and content mismatch")
		}
		subChannelProposal, err := fromSubChannelProposal(msg)
		if err != nil {
			return err
		}
		grpcMsg = &Envelope_SubChannelProposalMsg{
			SubChannelProposalMsg: subChannelProposal,
		}
	case wire.VirtualChannelProposal:
		msg, ok := env.Msg.(*client.VirtualChannelProposalMsg)
		if !ok {
			return errors.New("message type and content mismatch")
		}
		virtualChannelProposal, err := fromVirtualChannelProposal(msg)
		if err != nil {
			return err
		}
		grpcMsg = &Envelope_VirtualChannelProposalMsg{
			VirtualChannelProposalMsg: virtualChannelProposal,
		}
	case wire.LedgerChannelProposalAcc:
		msg, ok := env.Msg.(*client.LedgerChannelProposalAccMsg)
		if !ok {
			return errors.New("message type and content mismatch")
		}
		ledgerChannelProposalAcc, err := fromLedgerChannelProposalAcc(msg)
		if err != nil {
			return err
		}
		grpcMsg = &Envelope_LedgerChannelProposalAccMsg{
			LedgerChannelProposalAccMsg: ledgerChannelProposalAcc,
		}
	case wire.SubChannelProposalAcc:
		msg, ok := env.Msg.(*client.SubChannelProposalAccMsg)
		if !ok {
			return errors.New("message type and content mismatch")
		}
		grpcMsg = &Envelope_SubChannelProposalAccMsg{
			SubChannelProposalAccMsg: fromSubChannelProposalAcc(msg),
		}
	case wire.VirtualChannelProposalAcc:
		msg, ok := env.Msg.(*client.VirtualChannelProposalAccMsg)
		if !ok {
			return errors.New("message type and content mismatch")
		}
		virtualChannelProposalAcc, err := fromVirtualChannelProposalAcc(msg)
		if err != nil {
			return err
		}
		grpcMsg = &Envelope_VirtualChannelProposalAccMsg{
			VirtualChannelProposalAccMsg: virtualChannelProposalAcc,
		}
	case wire.ChannelProposalRej:
		msg, ok := env.Msg.(*client.ChannelProposalRejMsg)
		if !ok {
			return errors.New("message type and content mismatch")
		}
		grpcMsg = &Envelope_ChannelProposalRejMsg{
			ChannelProposalRejMsg: fromChannelProposalRej(msg),
		}
	}

	protoEnv := Envelope{
		Sender:    sender,
		Recipient: recipient,
		Msg:       grpcMsg,
	}

	data, err := proto.Marshal(&protoEnv)
	if err != nil {
		return errors.Wrap(err, "marshalling envelope")
	}

	if err := binary.Write(w, binary.BigEndian, uint16(len(data))); err != nil {
		return errors.Wrap(err, "writing length to wire")
	}

	_, err = w.Write(data)
	return errors.Wrap(err, "writing data to wire")
}

// Decode decodes an envelope from the wire using perunio encoding format.
func (Serializer) Decode(r io.Reader) (*wire.Envelope, error) { // nolint: cyclop
	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, errors.Wrap(err, "reading length from wire")
	}
	data := make([]byte, length)
	if _, err := r.Read(data); err != nil {
		return nil, errors.Wrap(err, "reading data from wire")
	}
	var protoEnv Envelope
	if err := proto.Unmarshal(data, &protoEnv); err != nil {
		return nil, errors.Wrap(err, "unmarshalling envelope")
	}

	env := wire.Envelope{}
	env.Sender = wire.NewAddress()
	if err := env.Sender.UnmarshalBinary(protoEnv.Sender); err != nil {
		return nil, errors.Wrap(err, "unmarshalling sender address")
	}
	env.Recipient = wire.NewAddress()
	if err := env.Recipient.UnmarshalBinary(protoEnv.Recipient); err != nil {
		return nil, errors.Wrap(err, "unmarshalling recipient address")
	}

	var err error
	switch protoEnv.Msg.(type) {
	case *Envelope_PingMsg:
		env.Msg = &wire.PingMsg{
			PingPongMsg: wire.PingPongMsg{
				Created: time.Unix(0, protoEnv.GetPingMsg().Created),
			},
		}
	case *Envelope_PongMsg:
		env.Msg = &wire.PongMsg{
			PingPongMsg: wire.PingPongMsg{
				Created: time.Unix(0, protoEnv.GetPongMsg().Created),
			},
		}
	case *Envelope_ShutdownMsg:
		env.Msg = &wire.ShutdownMsg{
			Reason: protoEnv.GetShutdownMsg().Reason,
		}
	case *Envelope_AuthResponseMsg:
		env.Msg = &wire.AuthResponseMsg{}
	case *Envelope_LedgerChannelProposalMsg:
		env.Msg, err = toLedgerChannelProposal(protoEnv.GetLedgerChannelProposalMsg())
	case *Envelope_SubChannelProposalMsg:
		env.Msg, err = toSubChannelProposal(protoEnv.GetSubChannelProposalMsg())
	case *Envelope_VirtualChannelProposalMsg:
		env.Msg, err = toVirtualChannelProposal(protoEnv.GetVirtualChannelProposalMsg())
	case *Envelope_LedgerChannelProposalAccMsg:
		env.Msg, err = toLedgerChannelProposalAcc(protoEnv.GetLedgerChannelProposalAccMsg())
	case *Envelope_SubChannelProposalAccMsg:
		env.Msg = toSubChannelProposalAcc(protoEnv.GetSubChannelProposalAccMsg())
	case *Envelope_VirtualChannelProposalAccMsg:
		env.Msg, err = toVirtualChannelProposalAcc(protoEnv.GetVirtualChannelProposalAccMsg())
	case *Envelope_ChannelProposalRejMsg:
		env.Msg = toChannelProposalRej(protoEnv.GetChannelProposalRejMsg())
	}

	return &env, err
}
