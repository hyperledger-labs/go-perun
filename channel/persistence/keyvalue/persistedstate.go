// Copyright 2020 - See NOTICE file for copyright holders.
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

	"perun.network/go-perun/channel"
	perunio "perun.network/go-perun/pkg/io"
)

var _ perunio.Encoder = PersistedState{}
var _ perunio.Decoder = (*PersistedState)(nil)

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
