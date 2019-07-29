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
const maxBigIntLength uint8 = 128

// BigInt is a serializable big integer.
type BigInt big.Int

// Decode reads a big.Int from the given stream.
func (b *BigInt) Decode(reader io.Reader) error {
	// Read length
	var lengthData = make([]byte, 1)
	n, err := reader.Read(lengthData)

	if err != nil {
		return errors.Wrap(err, "failed to decode length of big.Int")
	}
	if n != 1 {
		return errors.New("failed to decode length of big.Int")
	}
	var length = uint8(lengthData[0])

	if length > maxBigIntLength || length < 0 {
		return errors.New("big.Int to big too decode")
	}

	bytes := make([]byte, length)
	n, err = io.ReadFull(reader, bytes)

	if err != nil {
		return errors.Wrap(err, "failed to read []byte in big.Int")
	}
	if n != int(length) {
		return errors.New("failed to read []byte in big.Int")
	}
	tmp := new(big.Int)
	*b = BigInt(*tmp.SetBytes(bytes))

	return nil
}

// Encode writes a big.Int to the stream.
func (b BigInt) Encode(writer io.Writer) error {
	integer := big.Int(b)
	bytes := integer.Bytes()
	// Dont cast it to SizeType here, otherwise it can overflow
	length := len(bytes)

	// 255 hardcoded because we serialize as uint8
	if length > int(maxBigIntLength) || length > 255 || length < 0 {
		return errors.New("big.Int to big too encode")
	}
	// Write length
	n, err := writer.Write([]byte{uint8(length)})

	if err != nil {
		return errors.Wrap(err, "failed to write length")
	}
	if n != 1 {
		return errors.New("failed to write length")
	}

	// Write bytes
	n, err = writer.Write(bytes)
	if n != length {
		return errors.New("failed to write big.Int")
	}
	return errors.Wrap(err, "failed to write big.Int")
}
