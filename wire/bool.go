// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"io"

	"github.com/pkg/errors"
)

// Bool is a serializable network boolean.
type Bool bool

func (b *Bool) Decode(reader io.Reader) error {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return errors.Wrap(err, "failed to read bool")
	}
	*b = Bool(buf[0] != 0)
	return nil
}

func (b Bool) Encode(writer io.Writer) error {
	buf := []byte{0}
	if b {
		buf[0] = 1
	}

	if _, err := writer.Write(buf); err != nil {
		return errors.Wrap(err, "failed to write bool")
	}
	return nil
}
