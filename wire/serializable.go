// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire // import "perun.network/go-perun/wire"

import (
	"io"

	"github.com/pkg/errors"
)

// Serializable objects can be serialised into and from streams.
type Serializable interface {
	// Decode reads an object from a stream.
	// If the stream fails, the underlying error is returned.
	// Returns an InvalidEncodingError if the stream's data is invalid.
	Decode(io.Reader) error
	// Encode writes an object to a stream.
	// If the stream fails, the underyling error is returned.
	Encode(io.Writer) error
}

func Encode(writer io.Writer, values ...Serializable) error {
	for _, v := range values {
		if err := v.Encode(writer); err != nil {
			return err
		}
	}

	return nil
}

func Decode(reader io.Reader, values ...Serializable) error {
	for _, v := range values {
		if err := v.Decode(reader); err != nil {
			return err
		}
	}

	return nil
}

// Int32 is a serializable network 32 bit integer.
type Int32 int32

func (i32 *Int32) Decode(reader io.Reader) error {
	buf := [4]byte{}
	if _, err := reader.Read(buf[:]); err != nil {
		return errors.Wrap(err, "failed to read int32")
	}
	*i32 = Int32((int(buf[0]) << 24) | (int(buf[1]) << 16) | (int(buf[2]) << 8) | (int(buf[3])))
	return nil
}

func (i32 Int32) Encode(writer io.Writer) error {
	buf := [4]byte{byte(i32), byte(i32 >> 8), byte(i32 >> 16), byte(i32 >> 24)}
	if _, err := writer.Write(buf[:]); err != nil {
		return errors.Wrap(err, "failed to write int32")
	}
	return nil
}

// Int16 is a serializable network 16 bit integer.
type Int16 int16

func (i16 *Int16) Decode(reader io.Reader) error {
	buf := [2]byte{}
	if _, err := reader.Read(buf[:]); err != nil {
		return errors.Wrap(err, "failed to read int16")
	}
	*i16 = Int16(int(buf[0]) | (int(buf[1]) << 8))

	return nil
}

func (i16 Int16) Encode(writer io.Writer) error {
	buf := [2]byte{byte(i16), byte(i16 >> 8)}
	if _, err := writer.Write(buf[:]); err != nil {
		return errors.Wrap(err, "failed to write int16")
	}

	return nil
}

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
	buf := [1]byte{}
	if _, err := writer.Write(buf[:]); err != nil {
		return errors.Wrap(err, "failed to write bool")
	}
	return nil
}
