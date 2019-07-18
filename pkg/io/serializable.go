// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package io contains the serialization interfaces used by perun.
package io // import "perun.network/go-perun/pkg/io"

import (
	"io"

	"github.com/pkg/errors"
)

// Serializable objects can be serialized into and from streams.
type Serializable interface {
	// Decode reads an object from a stream.
	// If the stream fails, the underlying error is returned.
	// Returns an error if the stream's data is invalid.
	Decode(Reader) error
	// Encode writes an object to a stream.
	// If the stream fails, the underyling error is returned.
	Encode(Writer) error
}

// Encode encodes multiple serializable objects at once.
// If an error occurs, the index at which it occurs is also reported.
func Encode(writer Writer, values ...Serializable) error {
	for i, v := range values {
		if err := v.Encode(writer); err != nil {
			return errors.WithMessagef(err, "failed to encode %dth object (%T)", i, v)
		}
	}

	return nil
}

// Decode decodes multiple serializable objects at once.
// If an error occurs, the index at which it occurs is also reported.
func Decode(reader Reader, values ...Serializable) error {
	for i, v := range values {
		if err := v.Decode(reader); err != nil {
			return errors.WithMessagef(err, "failed to decode %dth object (%T)", i, v)
		}
	}

	return nil
}

// Writer exports io.Writer.
type Writer = io.Writer

// Reader exports io.Reader.
type Reader = io.Reader

// ReadWriter exports io.ReadWriter.
type ReadWriter = io.ReadWriter

// ReadWriteCloser exports io.ReadWriteCloser.
type ReadWriteCloser = io.ReadWriteCloser

// Closer exports io.Closer.
type Closer = io.Closer
