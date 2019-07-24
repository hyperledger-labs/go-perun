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
type BigInt big.Int

// Decode reads a big.Int from the given stream.
func (b *BigInt) Decode(reader io.Reader) error {
	// Read length
	var length Int16
	if err := length.Decode(reader); err != nil {
		return errors.Wrap(err, "failed to decode length of big.Int")
	}

	if length > maxBigIntLength {
		return errors.New("big.Int to big too decode")
	}

	bytes := make([]byte, length)
	if _, err := io.ReadFull(reader, bytes); err != nil {
		return errors.Wrap(err, "failed to read []byte in big.Int")
	}
	tmp := new(big.Int)
	*b = BigInt(*tmp.SetBytes(bytes))
	return nil
}

// Encode writes a big.Int to the stream.
func (b BigInt) Encode(writer io.Writer) error {
	integer := big.Int(b)
	bytes := integer.Bytes()
	// Write length
	length := Int16(len(bytes))
	if length > maxBigIntLength {
		return errors.New("big.Int to big too encode")
	}
	if err := length.Encode(writer); err != nil {
		return errors.Wrap(err, "failed to write length")
	}
	// Write bytes
	if _, err := writer.Write(bytes); err != nil {
		return errors.Wrap(err, "failed to write big.Int")
	}
	return nil
}
