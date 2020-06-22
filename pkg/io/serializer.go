// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package io

import (
	"io"
)

type (
	// Serializer objects can be serialized into and from streams.
	Serializer interface {
		Encoder
		Decoder
	}

	// An Encoder can encode itself into a stream.
	Encoder interface {
		// Encode writes itself to a stream.
		// If the stream fails, the underlying error is returned.
		Encode(io.Writer) error
	}

	// A Decoder can decode itself from a stream.
	Decoder interface {
		// Decode reads an object from a stream.
		// If the stream fails, the underlying error is returned.
		Decode(io.Reader) error
	}
)
