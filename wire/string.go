// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

// encodeString writes the length as an uint16 and then the string itself to the io.Writer.
func encodeString(w io.Writer, s string) error {
	l := uint16(len(s))
	if int(l) != len(s) {
		return errors.Errorf("string length exceeded: %d", len(s))
	}

	if err := binary.Write(w, byteOrder, l); err != nil {
		return errors.Wrap(err, "failed to write string length")
	}

	// Early exit. Plus, io.WriteString will complain about a closed io.Writer
	// even if there is nothing left to write
	if l == 0 {
		return nil
	}

	_, err := io.WriteString(w, s)
	return errors.Wrap(err, "failed to write string")
}

// decodeString reads the length as uint16 and the the string itself from the io.Reader.
func decodeString(r io.Reader, s *string) error {
	var l uint16
	if err := binary.Read(r, byteOrder, &l); err != nil {
		return errors.Wrap(err, "failed to read string length")
	}

	buf := make([]byte, l)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return errors.Wrap(err, "failed to read string")
	}
	*s = string(buf)
	return nil
}
