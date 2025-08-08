// Copyright 2019 - See NOTICE file for copyright holders.
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

package perunio

import (
	"io"
	"math/big"

	"github.com/pkg/errors"
)

// MaxBigIntLength defines the maximum length of a big integer.
// 1024bit -> 128 bytes.
const MaxBigIntLength = 128

// BigInt is a serializer big integer.
type BigInt struct {
	*big.Int
}

// Decode reads a big.Int from the given stream.
func (b *BigInt) Decode(reader io.Reader) error {
	// Read length
	lengthData := make([]byte, 1)
	_, err := reader.Read(lengthData)
	if err != nil {
		return errors.Wrap(err, "failed to decode length of big.Int")
	}

	length := lengthData[0]
	if length > MaxBigIntLength {
		return errors.New("big.Int too big to decode")
	}

	bytes := make([]byte, length)
	n, err := io.ReadFull(reader, bytes)
	if err != nil {
		return errors.Wrapf(err, "failed to read bytes for big.Int, read %d/%d", n, length)
	}

	if b.Int == nil {
		b.Int = new(big.Int)
	}

	b.SetBytes(bytes)

	return nil
}

// Encode writes a big.Int to the stream.
func (b BigInt) Encode(writer io.Writer) error {
	if b.Int == nil {
		panic("logic error: tried to encode nil big.Int")
	}

	if b.Sign() == -1 {
		panic("encoding of negative big.Int not implemented")
	}

	bytes := b.Bytes()
	length := len(bytes)
	// we serialize the length as uint8
	if length > MaxBigIntLength {
		return errors.New("big.Int too big to encode")
	}

	// Write length
	_, err := writer.Write([]byte{uint8(length)})
	if err != nil {
		return errors.Wrap(err, "failed to write length")
	}

	if length == 0 {
		return nil
	}

	// Write bytes
	n, err := writer.Write(bytes)

	return errors.Wrapf(err, "failed to write big.Int, wrote %d bytes of %d", n, length)
}
