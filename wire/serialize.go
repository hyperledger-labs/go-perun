// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"io"
	"math/big"
	"time"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
)

// Encode encodes multiple primitive values into a writer.
// All passed values must be copies, not references.
func Encode(writer io.Writer, values ...interface{}) (err error) {
	for i, value := range values {
		switch v := value.(type) {
		case bool:
			err = Bool(v).Encode(writer)
		case int16:
			err = Int16(v).Encode(writer)
		case uint16:
			err = Int16(v).Encode(writer)
		case int32:
			err = Int32(v).Encode(writer)
		case uint32:
			err = Int32(v).Encode(writer)
		case int64:
			err = Int64(v).Encode(writer)
		case uint64:
			err = Int64(v).Encode(writer)
		case time.Time:
			err = FromTime(v).Encode(writer)
		case *big.Int:
			err = BigInt{v}.Encode(writer)
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
		case *bool:
			var d Bool
			err = d.Decode(reader)
			*v = bool(d)
		case *int16:
			var d Int16
			err = d.Decode(reader)
			*v = int16(d)
		case *uint16:
			var d Int16
			err = d.Decode(reader)
			*v = uint16(d)
		case *int32:
			var d Int32
			err = d.Decode(reader)
			*v = int32(d)
		case *uint32:
			var d Int32
			err = d.Decode(reader)
			*v = uint32(d)
		case *int64:
			var d Int64
			err = d.Decode(reader)
			*v = int64(d)
		case *uint64:
			var d Int64
			err = d.Decode(reader)
			*v = uint64(d)
		case *time.Time:
			var d Time
			err = d.Decode(reader)
			*v = d.Time()
		case **big.Int:
			var d BigInt
			err = d.Decode(reader)
			*v = d.Int
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
