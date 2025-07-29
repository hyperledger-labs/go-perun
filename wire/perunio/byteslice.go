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

	"github.com/pkg/errors"
)

// ByteSlice is a serializer byte slice.
type ByteSlice []byte

var _ Serializer = (*ByteSlice)(nil)

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
