// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package io

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

var byteOrder = binary.LittleEndian

// Encode encodes multiple primitive values into a writer.
// All passed values must be copies, not references.
func Encode(writer io.Writer, values ...interface{}) (err error) {
	for i, value := range values {
		switch v := value.(type) {
		case bool, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
			err = binary.Write(writer, byteOrder, v)
		case time.Time:
			err = binary.Write(writer, byteOrder, v.UnixNano())
		case *big.Int:
			err = BigInt{v}.Encode(writer)
		case [32]byte:
			_, err = writer.Write(v[:])
		case []byte:
			err = ByteSlice(v).Encode(writer)
		case string:
			err = encodeString(writer, v)
		default:
			if enc, ok := value.(Encoder); ok {
				err = enc.Encode(writer)
			} else {
				panic(fmt.Sprintf("perunio.Encode(): Invalid type %T", v))
			}
		}

		if err != nil {
			return errors.WithMessagef(err, "failed to encode %dth value of type %T", i, value)
		}
	}

	return nil
}

// Decode decodes multiple primitive values from a reader.
// All passed values must be references, not copies.
func Decode(reader io.Reader, values ...interface{}) (err error) {
	for i, value := range values {
		switch v := value.(type) {
		case *bool, *int8, *uint8, *int16, *uint16, *int32, *uint32, *int64, *uint64:
			err = binary.Read(reader, byteOrder, v)
		case *time.Time:
			var nsec int64
			err = binary.Read(reader, byteOrder, &nsec)
			*v = time.Unix(0, nsec)
		case **big.Int:
			var d BigInt
			err = d.Decode(reader)
			*v = d.Int
		case *[32]byte:
			_, err = io.ReadFull(reader, v[:])
		case *[]byte:
			d := ByteSlice(*v)
			err = d.Decode(reader)
		case *string:
			err = decodeString(reader, v)
		default:
			if dec, ok := value.(Decoder); ok {
				err = dec.Decode(reader)
			} else {
				panic(fmt.Sprintf("perunio.Decode(): Invalid type %T", v))
			}
		}

		if err != nil {
			return errors.WithMessagef(err, "failed to decode %dth value of type %T", i, value)
		}
	}

	return nil
}
