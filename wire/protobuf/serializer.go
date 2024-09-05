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
	"fmt"
	"io"
	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wire"
)

type serializer struct{}

// Serializer returns a protobuf serializer.
func Serializer() wire.EnvelopeSerializer {
	return serializer{}
}

// Encode encodes an envelope from the reader using protocol buffers
// serialization format.
func (serializer) Encode(w io.Writer, env *wire.Envelope) (err error) { //nolint: funlen, cyclop
	protoEnv := &Envelope{}

	switch msg := env.Msg.(type) {
	case *wire.PingMsg:
		protoEnv.Msg = fromPingMsg(msg)
	case *wire.PongMsg:
		protoEnv.Msg = fromPongMsg(msg)
	case *wire.ShutdownMsg:
		protoEnv.Msg = fromShutdownMsg(msg)
	case *wire.AuthResponseMsg:
		protoEnv.Msg = fromAuthResponseMsg(msg)
	case *client.LedgerChannelProposalMsg:
		protoEnv.Msg, err = FromLedgerChannelProposalMsg(msg)
	case *client.SubChannelProposalMsg:
		protoEnv.Msg, err = FromSubChannelProposalMsg(msg)
	case *client.VirtualChannelProposalMsg:
		protoEnv.Msg, err = FromVirtualChannelProposalMsg(msg)
	case *client.LedgerChannelProposalAccMsg:
		protoEnv.Msg, err = FromLedgerChannelProposalAccMsg(msg)
	case *client.SubChannelProposalAccMsg:
		protoEnv.Msg = FromSubChannelProposalAccMsg(msg)
	case *client.VirtualChannelProposalAccMsg:
		protoEnv.Msg, err = FromVirtualChannelProposalAccMsg(msg)
	case *client.ChannelProposalRejMsg:
		protoEnv.Msg = FromChannelProposalRejMsg(msg)
	case *client.ChannelUpdateMsg:
		protoEnv.Msg, err = FromChannelUpdateMsg(msg)
	case *client.VirtualChannelFundingProposalMsg:
		protoEnv.Msg, err = FromVirtualChannelFundingProposalMsg(msg)
	case *client.VirtualChannelSettlementProposalMsg:
		protoEnv.Msg, err = FromVirtualChannelSettlementProposalMsg(msg)
	case *client.ChannelUpdateAccMsg:
		protoEnv.Msg = FromChannelUpdateAccMsg(msg)
	case *client.ChannelUpdateRejMsg:
		protoEnv.Msg = FromChannelUpdateRejMsg(msg)
	case *client.ChannelSyncMsg:
		protoEnv.Msg, err = fromChannelSyncMsg(msg)
	default:
		//nolint: goerr113  // We do not want to define this as constant error.
		err = fmt.Errorf("unknown message type: %T", msg)
	}
	if err != nil {
		return err
	}
	sender, recipient, err := marshalSenderRecipient(env)
	protoEnv.Sender, protoEnv.Recipient = sender, recipient
	if err != nil {
		return err
	}

	return writeEnvelope(w, protoEnv)
}

func marshalSenderRecipient(env *wire.Envelope) (*Address, *Address, error) {
	sender, err := FromWireAddr(env.Sender)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "marshalling sender address")
	}
	recipient, err := FromWireAddr(env.Recipient)
	return sender, recipient, errors.WithMessage(err, "marshalling recipient address")
}

func writeEnvelope(w io.Writer, env *Envelope) error {
	data, err := proto.Marshal(env)
	if err != nil {
		return errors.Wrap(err, "marshalling envelope")
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(data))); err != nil {
		return errors.Wrap(err, "writing length to wire")
	}
	_, err = w.Write(data)
	return errors.Wrap(err, "writing data to wire")
}

// Decode decodes an envelope from the reader, that was encoded using protocol
// buffers serialization format.
func (serializer) Decode(r io.Reader) (env *wire.Envelope, err error) { //nolint: funlen, cyclop
	env = &wire.Envelope{}

	protoEnv, err := readEnvelope(r)
	if err != nil {
		return nil, err
	}

	sender, recipient, err := unmarshalSenderRecipient(protoEnv)
	env.Sender, env.Recipient = sender, recipient
	if err != nil {
		return nil, err
	}

	switch protoMsg := protoEnv.Msg.(type) {
	case *Envelope_PingMsg:
		env.Msg = toPingMsg(protoMsg)
	case *Envelope_PongMsg:
		env.Msg = toPongMsg(protoMsg)
	case *Envelope_ShutdownMsg:
		env.Msg = toShutdownMsg(protoMsg)
	case *Envelope_AuthResponseMsg:
		env.Msg = toAuthResponseMsg(protoMsg)
	case *Envelope_LedgerChannelProposalMsg:
		env.Msg, err = ToLedgerChannelProposalMsg(protoMsg)
	case *Envelope_SubChannelProposalMsg:
		env.Msg, err = ToSubChannelProposalMsg(protoMsg)
	case *Envelope_VirtualChannelProposalMsg:
		env.Msg, err = ToVirtualChannelProposalMsg(protoMsg)
	case *Envelope_LedgerChannelProposalAccMsg:
		env.Msg, err = ToLedgerChannelProposalAccMsg(protoMsg)
	case *Envelope_SubChannelProposalAccMsg:
		env.Msg = ToSubChannelProposalAccMsg(protoMsg)
	case *Envelope_VirtualChannelProposalAccMsg:
		env.Msg, err = ToVirtualChannelProposalAccMsg(protoMsg)
	case *Envelope_ChannelProposalRejMsg:
		env.Msg = ToChannelProposalRejMsg(protoMsg)
	case *Envelope_ChannelUpdateMsg:
		env.Msg, err = ToChannelUpdateMsg(protoMsg)
	case *Envelope_VirtualChannelFundingProposalMsg:
		env.Msg, err = ToVirtualChannelFundingProposalMsg(protoMsg)
	case *Envelope_VirtualChannelSettlementProposalMsg:
		env.Msg, err = ToVirtualChannelSettlementProposalMsg(protoMsg)
	case *Envelope_ChannelUpdateAccMsg:
		env.Msg = ToChannelUpdateAccMsg(protoMsg)
	case *Envelope_ChannelUpdateRejMsg:
		env.Msg = ToChannelUpdateRejMsg(protoMsg)
	case *Envelope_ChannelSyncMsg:
		env.Msg, err = toChannelSyncMsg(protoMsg)
	default:
		//nolint: goerr113  // We do not want to define this as constant error.
		err = fmt.Errorf("unknown message type: %T", protoMsg)
	}

	return env, err
}

func readEnvelope(r io.Reader) (*Envelope, error) {
	var size uint16
	if err := binary.Read(r, binary.BigEndian, &size); err != nil {
		return nil, errors.Wrap(err, "reading size of data from wire")
	}
	data := make([]byte, size)
	if _, err := r.Read(data); err != nil {
		return nil, errors.Wrap(err, "reading data from wire")
	}
	var protoEnv Envelope
	return &protoEnv, errors.Wrap(proto.Unmarshal(data, &protoEnv), "unmarshalling envelope")
}

func unmarshalSenderRecipient(protoEnv *Envelope) (map[wallet.BackendID]wire.Address, map[wallet.BackendID]wire.Address, error) {
	sender, err := ToWireAddr(protoEnv.Sender)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unmarshalling sender address")
	}
	recipient, err1 := ToWireAddr(protoEnv.Recipient)
	if err1 != nil {
		return nil, nil, errors.Wrap(err, "unmarshalling recipient address")
	}
	return sender, recipient, nil
}
