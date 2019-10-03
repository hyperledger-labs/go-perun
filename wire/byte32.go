// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"io"

	"github.com/pkg/errors"
)

// Byte32 is a serializable 32-byte-array
type Byte32 [32]byte

func (b *Byte32) Decode(r io.Reader) error {
	_, err := io.ReadFull(r, b[:])
	return errors.Wrap(err, "error decoding [32]byte")
}

func (b Byte32) Encode(w io.Writer) error {
	_, err := w.Write(b[:])
	return errors.Wrap(err, "error encoding [32]byte")
}
