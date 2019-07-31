// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"io"

	"github.com/pkg/errors"
)

// ByteSlice is a serializable byte slice.
type ByteSlice []byte

// Decode reads a byte slice from the given stream.
// Decode reads exactly len(b) bytes.
// This means the caller has to specify how many bytes he wants to read.
func (b *ByteSlice) Decode(reader io.Reader) (err error) {
	n, err := reader.Read(*b)
	for n < len(*b) && err == nil {
		var nn int
		nn, err = reader.Read((*b)[n:])
		n += nn
	}
	return errors.Wrap(err, "failed to read []byte")
}

// Encode writes len(b) to the stream.
func (b ByteSlice) Encode(writer io.Writer) error {
	_, err := writer.Write(b)
	return errors.Wrap(err, "failed to write []byte")
}
