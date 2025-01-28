// Copyright 2025 - See NOTICE file for copyright holders.
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

package keyvalue

import (
	"io"

	"perun.network/go-perun/wallet"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire/perunio"
)

var (
	_ perunio.Encoder = PersistedState{}
	_ perunio.Decoder = (*PersistedState)(nil)
)

// PersistedState is a helper struct to allow for de-/encoding of empty states.
type PersistedState struct {
	State **channel.State
}

// Encode writes itself to a stream.
// If the stream fails, the underlying error is returned.
func (s PersistedState) Encode(w io.Writer) error {
	if (*s.State) == nil {
		return nil
	}
	return (*s.State).Encode(w)
}

// Decode reads a channel.State from an `io.Reader`.
func (s *PersistedState) Decode(r io.Reader) error {
	*s.State = new(channel.State)
	return (*s.State).Decode(r)
}

type (
	optChannelIDEnc struct {
		ID *map[wallet.BackendID]channel.ID
	}

	optChannelIDDec struct {
		ID **map[wallet.BackendID]channel.ID
	}
)

func (id optChannelIDEnc) Encode(w io.Writer) error {
	if id.ID != nil {
		return perunio.Encode(w, true, channel.IDMap(*id.ID))
	}
	return perunio.Encode(w, false)
}

func (id optChannelIDDec) Decode(r io.Reader) error {
	var exists bool
	if err := perunio.Decode(r, &exists); err != nil {
		return err
	}
	if exists {
		*id.ID = new(map[wallet.BackendID]channel.ID)
		return perunio.Decode(r, (*channel.IDMap)(*id.ID))
	}
	*id.ID = nil
	return nil
}
