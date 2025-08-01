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
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/pkg/errors"
)

// encodeString writes the length as an uint16 and then the string itself to the io.Writer.
func encodeString(w io.Writer, s string) error {
	l := len(s)
	if l > int(math.MaxUint16) {
		return fmt.Errorf("string too long: %d", l)
	}
	ul := uint16(l)
	if int(ul) != len(s) {
		return errors.Errorf("string length exceeded: %d", len(s))
	}

	err := binary.Write(w, byteOrder, ul)
	if err != nil {
		return errors.Wrap(err, "failed to write string length")
	}

	// Early exit. Plus, io.WriteString will complain about a closed io.Writer
	// even if there is nothing left to write
	if l == 0 {
		return nil
	}

	_, err = io.WriteString(w, s)

	return errors.Wrap(err, "failed to write string")
}

// decodeString reads the length as uint16 and the string itself from the io.Reader.
func decodeString(r io.Reader, s *string) error {
	var l uint16
	err := binary.Read(r, byteOrder, &l)
	if err != nil {
		return errors.Wrap(err, "failed to read string length")
	}

	buf := make([]byte, l)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return errors.Wrap(err, "failed to read string")
	}
	*s = string(buf)
	return nil
}
