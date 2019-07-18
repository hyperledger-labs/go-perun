// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"github.com/pkg/errors"
	"io"
)

// Bool is a serializable network boolean.
type Bool bool

func (b *Bool) Decode(reader io.Reader) error {
	buf := [1]byte{}
	if _, err := reader.Read(buf[:]); err != nil {
		return errors.Wrap(err, "failed to read bool")
	}
	*b = Bool(buf[0] != 0)
	return nil
}

func (b Bool) Encode(writer io.Writer) error {
	var v byte
	if b {
		v = 1
	} else {
		v = 0
	}
	buf := [1]byte{v}
	if _, err := writer.Write(buf[:]); err != nil {
		return errors.Wrap(err, "failed to write bool")
	}
	return nil
}
