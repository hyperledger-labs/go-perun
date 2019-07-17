// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"github.com/pkg/errors"
	"io"
)

// Int64 is a serializable 64-bit integer.
type Int64 int64

func (i64 *Int64) Decode(reader io.Reader) error {
	buf := [8]byte{}
	if _, err := reader.Read(buf[:]); err != nil {
		return errors.Wrap(err, "failed to read i64")
	}

	var u64 uint64
	for i := 0; i < 8; i++ {
		u64 |= uint64(buf[i]) << (uint64(8) * uint64(i))
	}

	*i64 = Int64(int64(u64))
	return nil
}

func (i64 Int64) Encode(writer io.Writer) error {
	buf := [8]byte{}
	for i := 0; i < 8; i++ {
		buf[i] = byte(uint64(i64) >> (uint64(8) * uint64(i)))
	}

	if _, err := writer.Write(buf[:]); err != nil {
		return errors.Wrap(err, "failed to write i64")
	}

	return nil
}
