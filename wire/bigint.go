// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"io"
	"math/big"

	"github.com/pkg/errors"
)

// maxBigIntLength defines the maximum length of a big integer.
// Default: 1024bit -> 128 bytes
const maxBigIntLength = 128

// BigInt is a serializable big integer.
type BigInt struct {
	*big.Int
}

// Decode reads a big.Int from the given stream.
func (b *BigInt) Decode(reader io.Reader) error {
	// Read length
	var lengthData = make([]byte, 1)
	_, err := reader.Read(lengthData)
	if err != nil {
		return errors.Wrap(err, "failed to decode length of big.Int")
	}

	var length = uint8(lengthData[0])
	if length > maxBigIntLength {
		return errors.New("big.Int too big to decode")
	}

	if length == 0 {
		*b = BigInt{big.NewInt(0)}
		return nil
	}

	bytes := make([]byte, length)
	_, err = io.ReadFull(reader, bytes)
	if err != nil {
		return errors.Wrap(err, "failed to read []byte in big.Int")
	}

	b.Int = new(big.Int).SetBytes(bytes)
	return nil
}

// Encode writes a big.Int to the stream.
func (b BigInt) Encode(writer io.Writer) error {
	bytes := b.Bytes()
	// Dont cast it to SizeType here, otherwise it can overflow
	length := len(bytes)

	// 255 hardcoded because we serialize as uint8
	if uint8(length) > maxBigIntLength {
		return errors.New("big.Int too big to encode")
	}

	// Write length
	n, err := writer.Write([]byte{uint8(length)})

	if err != nil {
		return errors.Wrap(err, "failed to write length")
	}

	if length == 0 {
		return nil
	}

	// Write bytes
	n, err = writer.Write(bytes)
	return errors.Wrapf(err, "failed to write big.Int, wrote %d bytes of %d", n, length)
}
