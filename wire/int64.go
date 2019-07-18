// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

// Int64 is a serializable 64-bit integer.
type Int64 int64

func (i64 *Int64) Decode(reader io.Reader) error {
	buf := [8]byte{}
	if _, err := reader.Read(buf[:]); err != nil {
		return errors.Wrap(err, "failed to read i64")
	}

	*i64 = Int64(binary.LittleEndian.Uint64(buf[:]))
	return nil
}

func (i64 Int64) Encode(writer io.Writer) error {
	buf := [8]byte{}
	binary.LittleEndian.PutUint64(buf[:], uint64(i64))

	if _, err := writer.Write(buf[:]); err != nil {
		return errors.Wrap(err, "failed to write i64")
	}

	return nil
}
