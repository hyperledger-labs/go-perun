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
func (b *ByteSlice) Decode(reader io.Reader) error {
	if _, err := io.ReadFull(reader, *b); err != nil {
		return errors.Wrap(err, "failed to read []byte")
	}
	return nil
}

// Encode writes len(b) to the stream.
func (b ByteSlice) Encode(writer io.Writer) error {
	if _, err := writer.Write(b); err != nil {
		return errors.Wrap(err, "failed to write []byte")
	}
	return nil
}
