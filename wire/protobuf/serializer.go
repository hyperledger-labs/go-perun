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
func (Serializer) Encode(w io.Writer, env *wire.Envelope) error {
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
func (Serializer) Decode(r io.Reader) (*wire.Envelope, error) {
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
	}

	return &env, nil

}
