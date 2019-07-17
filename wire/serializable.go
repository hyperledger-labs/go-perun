// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package wire contains the network serialization code for primitive types.
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

// Encode encodes multiple serializable objects at once.
// If an error occurs, the index at which it occurs is also reported.
func Encode(writer io.Writer, values ...Serializable) error {
	for i, v := range values {
		if err := v.Encode(writer); err != nil {
			return errors.WithMessagef(err, "failed to encode %dth object", i)
		}
	}

	return nil
}

// Decode decodes multiple serializable objects at once.
// If an error occurs, the index at which it occurs is also reported.
func Decode(reader io.Reader, values ...Serializable) error {
	for i, v := range values {
		if err := v.Decode(reader); err != nil {
			return errors.WithMessagef(err, "failed to decode %dth object", i)
		}
	}

	return nil
}
