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
	if err := perunio.Encode(w, env.Sender, env.Recipient); err != nil {
		return err
	}
	return wire.EncodeMsg(env.Msg, w)
}

// Decode decodes an envelope from the wire using perunio encoding format.
func (serializer) Decode(r io.Reader) (env *wire.Envelope, err error) {
	env = &wire.Envelope{}
	env.Sender = wire.NewAddress()
	if err = perunio.Decode(r, env.Sender); err != nil {
		return env, errors.WithMessage(err, "decoding sender address")
	}
	env.Recipient = wire.NewAddress()
	if err = perunio.Decode(r, env.Recipient); err != nil {
		return env, errors.WithMessage(err, "decoding recipient address")
	}
	env.Msg, err = wire.DecodeMsg(r)
	return env, err
}
