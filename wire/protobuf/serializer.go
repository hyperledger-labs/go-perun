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

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"perun.network/go-perun/wire"
)

type serializer struct{}

func init() {
	wire.SetEnvelopeSerializer(serializer{})
}

// Encode encodes an envelope from the reader using protocol buffers
// serialization format.
func (serializer) Encode(w io.Writer, env *wire.Envelope) (err error) {
	protoEnv := &Envelope{}

	switch msg := env.Msg.(type) {
	case *wire.PingMsg:
		protoEnv.Msg = fromPingMsg(msg)
	case *wire.PongMsg:
		protoEnv.Msg = fromPongMsg(msg)
	case *wire.ShutdownMsg:
		protoEnv.Msg = fromShutdownMsg(msg)
	}

	protoEnv.Sender, protoEnv.Recipient, err = marshalSenderRecipient(env)
	if err != nil {
		return err
	}

	return writeEnvelope(w, protoEnv)
}

func marshalSenderRecipient(env *wire.Envelope) ([]byte, []byte, error) {
	sender, err := env.Sender.MarshalBinary()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "marshalling sender address")
	}
	recipient, err := env.Recipient.MarshalBinary()
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
func (serializer) Decode(r io.Reader) (env *wire.Envelope, err error) {
	env = &wire.Envelope{}

	protoEnv, err := readEnvelope(r)
	if err != nil {
		return nil, err
	}

	env.Sender, env.Recipient, err = unmarshalSenderRecipient(protoEnv)
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
	}

	return env, nil
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

func unmarshalSenderRecipient(protoEnv *Envelope) (wire.Address, wire.Address, error) {
	sender := wire.NewAddress()
	if err := sender.UnmarshalBinary(protoEnv.Sender); err != nil {
		return nil, nil, errors.Wrap(err, "unmarshalling sender address")
	}
	recipient := wire.NewAddress()
	err := recipient.UnmarshalBinary(protoEnv.Recipient)
	return sender, recipient, errors.Wrap(err, "unmarshalling recipient address")
}
