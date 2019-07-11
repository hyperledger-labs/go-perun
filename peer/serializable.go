// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"io"
)

// Serializable objects can be serialised into and from streams.
type Serializable interface {
	// Decode reads an object from a stream.
	// If the stream fails, the underlying error is returned.
	// Returns an InvalidEncodingError if the stream's data is invalid.
	Decode(io.Reader) error
	// Encode writes an objeect to a stream.
	// If the stream fails, the underyling error is returned.
	Encode(io.Writer) error
}

// InvalidEncodingError represents an error that occurred during decoding
// due to invalid encoded data.
type InvalidEncodingError struct {
	Reason string
}

func (e *InvalidEncodingError) Error() string {
	return "Invalid encoding: " + e.Reason
}
