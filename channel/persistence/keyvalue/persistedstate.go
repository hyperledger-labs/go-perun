// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

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
