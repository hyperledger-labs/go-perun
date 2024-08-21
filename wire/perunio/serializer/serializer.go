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

package serializer

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/perunio"
)

// Serializer returns a perunio serializer.
func Serializer() wire.EnvelopeSerializer {
	return serializer{}
}

type serializer struct{}

// Encode encodes the envelope into the wire using perunio encoding format.
func (serializer) Encode(w io.Writer, env *wire.Envelope) error {
	if err := perunio.Encode(w, wire.AddressDecMap(env.Sender), wire.AddressDecMap(env.Recipient)); err != nil {
		return err
	}
	return wire.EncodeMsg(env.Msg, w)
}

// Decode decodes an envelope from the wire using perunio encoding format.
func (serializer) Decode(r io.Reader) (env *wire.Envelope, err error) {
	env = &wire.Envelope{}
	err = perunio.Decode(r, (*wire.AddressDecMap)(&env.Sender))
	if err != nil {
		return env, errors.WithMessage(err, "decoding sender addresses")
	}

	// Decode the Recipient map
	err = perunio.Decode(r, (*wire.AddressDecMap)(&env.Recipient))
	if err != nil {
		return env, errors.WithMessage(err, "decoding recipient addresses")
	}

	// Decode the message
	env.Msg, err = wire.DecodeMsg(r)
	if err != nil {
		return env, errors.WithMessage(err, "decoding message")
	}

	return env, nil
}

// decodeAddressMap decodes a map[int]Address from the reader.
func decodeAddressMap(r io.Reader) (map[int]wire.Address, error) {
	var numEntries int32
	if err := binary.Read(r, binary.BigEndian, &numEntries); err != nil {
		return nil, errors.New("could not decode map length: " + err.Error())
	}

	addrMap := make(map[int]wire.Address, numEntries)
	for i := 0; i < int(numEntries); i++ {
		var idx int32
		if err := binary.Read(r, binary.BigEndian, &idx); err != nil {
			return nil, errors.New("could not decode map index: " + err.Error())
		}
		addrs := wire.NewAddress()
		if err := perunio.Decode(r, addrs); err != nil {
			return nil, errors.WithMessagef(err, "decoding %d-th address map entry", i)
		}
		addrMap[int(idx)] = addrs
	}

	return addrMap, nil
}
