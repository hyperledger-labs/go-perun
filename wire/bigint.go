// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"io"
	"math/big"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
)

// maxBigIntLength defines the maximum length of a big integer.
// 1024bit -> 128 bytes
const maxBigIntLength = 128

// BigInt is a serializable big integer.
type BigInt struct {
	*big.Int
}

// Decode reads a big.Int from the given stream.
func (b *BigInt) Decode(reader io.Reader) error {
	// Read length
	var lengthData = make([]byte, 1)
	if _, err := reader.Read(lengthData); err != nil {
		return errors.Wrap(err, "failed to decode length of big.Int")
	}

	var length = uint8(lengthData[0])
	if length > maxBigIntLength {
		return errors.New("big.Int too big to decode")
	}

	bytes := make([]byte, length)
	if n, err := io.ReadFull(reader, bytes); err != nil {
		return errors.Wrapf(err, "failed to read bytes for big.Int, read %d/%d", n, length)
	}

	if b.Int == nil {
		b.Int = new(big.Int)
	}
	b.Int.SetBytes(bytes)
	return nil
}

// Encode writes a big.Int to the stream.
func (b BigInt) Encode(writer io.Writer) error {
	if b.Int == nil {
		log.Panic("logic error: tried to encode nil big.Int")
	}
	if b.Int.Sign() == -1 {
		log.Panic("encoding of negative big.Int not implemented")
	}

	bytes := b.Bytes()
	length := len(bytes)
	// we serialize the length as uint8
	if length > maxBigIntLength {
		return errors.New("big.Int too big to encode")
	}

	// Write length
	if _, err := writer.Write([]byte{uint8(length)}); err != nil {
		return errors.Wrap(err, "failed to write length")
	}

	if length == 0 {
		return nil
	}

	// Write bytes
	n, err := writer.Write(bytes)
	return errors.Wrapf(err, "failed to write big.Int, wrote %d bytes of %d", n, length)
}
