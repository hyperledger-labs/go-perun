// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package wire

import (
	"io"

	"github.com/pkg/errors"
	perunio "perun.network/go-perun/pkg/io"
)

// ByteSlice is a serializer byte slice.
type ByteSlice []byte

var _ perunio.Serializer = (*ByteSlice)(nil)

// Encode writes len(b) bytes to the stream. Note that the length itself is not
// written to the stream.
func (b ByteSlice) Encode(w io.Writer) error {
	_, err := w.Write(b)
	return errors.Wrap(err, "failed to write []byte")
}

// Decode reads a byte slice from the given stream.
// Decode reads exactly len(b) bytes.
// This means the caller has to specify how many bytes he wants to read.
func (b *ByteSlice) Decode(r io.Reader) error {
	// This is almost the same as io.ReadFull, but it also fails on closed
	// readers.
	n, err := r.Read(*b)
	for n < len(*b) && err == nil {
		var nn int
		nn, err = r.Read((*b)[n:])
		n += nn
	}
	return errors.Wrap(err, "failed to read []byte")
}
