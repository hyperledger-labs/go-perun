// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"github.com/pkg/errors"
	"io"
)

// Time is a serializable network timestamp.
// It is a 64-bit unix timestamp, in nanoseconds.
type Time uint64

func (t *Time) Decode(reader io.Reader) error {
	var i64 Int64
	if err := i64.Decode(reader); err != nil {
		return errors.WithMessage(err, "failed to decode timestamp.")
	}
	*t = Time(i64)
	return nil
}

func (t Time) Encode(writer io.Writer) error {
	if err := Int64(t).Encode(writer); err != nil {
		return errors.WithMessage(err, "failed to encode timestamp.")
	}
	return nil
}
