// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package io

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

const uint16MaxValue = 0xFFFF

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
		case encoding.BinaryMarshaler:
			var data []byte
			data, err = v.MarshalBinary()
			if err != nil {
				return errors.WithMessage(err, "marshaling to byte array")
			}

			length := len(data)
			if length > uint16MaxValue {
				panic(fmt.Sprintf("lenth of marshaled data is %d, should be <= %d", len(data), uint16MaxValue))
			}
			err = binary.Write(writer, byteOrder, uint16(length))
			if err != nil {
				return errors.WithMessage(err, "writing length of marshalled data")
			}

			err = ByteSlice(data).Encode(writer)
		default:
			if enc, ok := value.(Encoder); ok {
				err = enc.Encode(writer)
			} else {
				panic(fmt.Sprintf("polyio.Encode(): Invalid type %T", v))
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
		case encoding.BinaryUnmarshaler:
			var length uint16
			err = binary.Read(reader, byteOrder, &length)
			if err != nil {
				return errors.WithMessage(err, "reading length of binary data")
			}

			var data ByteSlice = make([]byte, length)
			err = data.Decode(reader)
			if err != nil {
				return errors.WithMessage(err, "reading binary data")
			}

			err = v.UnmarshalBinary(data)
			err = errors.WithMessage(err, "unmarshaling binary data")
		default:
			if dec, ok := value.(Decoder); ok {
				err = dec.Decode(reader)
			} else {
				panic(fmt.Sprintf("polyio.Decode(): Invalid type %T", v))
			}
		}

		if err != nil {
			return errors.WithMessagef(err, "failed to decode %dth value of type %T", i, value)
		}
	}

	return nil
}
