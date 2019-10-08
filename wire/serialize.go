// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"encoding/binary"
	"io"
	"math/big"
	"time"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
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
			err = binary.Write(writer, byteOrder, FromTime(v).int64)
		case *big.Int:
			err = BigInt{v}.Encode(writer)
		case [32]byte:
			_, err = writer.Write(v[:])
		case []byte:
			err = ByteSlice(v).Encode(writer)
		default:
			log.Panicf("wire.Encode(): Invalid type %T", v)
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
			var d int64
			err = binary.Read(reader, byteOrder, &d)
			*v = Time{d}.Time()
		case **big.Int:
			var d BigInt
			err = d.Decode(reader)
			*v = d.Int
		case *[32]byte:
			_, err = io.ReadFull(reader, v[:])
		case *[]byte:
			d := ByteSlice(*v)
			err = d.Decode(reader)
		default:
			log.Panicf("wire.Decode(): Invalid type %T", v)
		}

		if err != nil {
			return errors.WithMessagef(err, "failed to decode %dth value of type %T", i, value)
		}
	}

	return nil
}
