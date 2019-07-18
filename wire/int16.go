// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

// Int16 is a serializable network 16 bit integer.
type Int16 int16

func (i16 *Int16) Decode(reader io.Reader) error {
	buf := [2]byte{}
	if _, err := reader.Read(buf[:]); err != nil {
		return errors.Wrap(err, "failed to read int16")
	}
	*i16 = Int16(binary.LittleEndian.Uint16(buf[:]))

	return nil
}

func (i16 Int16) Encode(writer io.Writer) error {
	buf := [2]byte{}
	binary.LittleEndian.PutUint16(buf[:], uint16(i16))

	if _, err := writer.Write(buf[:]); err != nil {
		return errors.Wrap(err, "failed to write int16")
	}

	return nil
}
