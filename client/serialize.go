// Copyright 2025 - See NOTICE file for copyright holders.
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

package client

import (
	"io"

	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire/perunio"
)

type (
	sliceLen          = uint16
	channelIDsWithLen []map[wallet.BackendID]channel.ID
	indexMapWithLen   []channel.Index
	indexMapsWithLen  [][]channel.Index
)

// Encode encodes the object to the writer.
func (a channelIDsWithLen) Encode(w io.Writer) (err error) {
	length := int32(len(a))
	if err := perunio.Encode(w, length); err != nil {
		return errors.WithMessage(err, "encoding array length")
	}
	for i, id := range a {
		idCopy := id
		if err := perunio.Encode(w, (*channel.IDMap)(&idCopy)); err != nil {
			return errors.WithMessagef(err, "encoding %d-th id array entry", i)
		}
	}
	return nil
}

// Decode decodes the object from the reader.
func (a *channelIDsWithLen) Decode(r io.Reader) (err error) {
	var mapLen int32
	if err := perunio.Decode(r, &mapLen); err != nil {
		return errors.WithMessage(err, "decoding array length")
	}
	*a = make([]map[wallet.BackendID]channel.ID, mapLen)
	for i := 0; i < int(mapLen); i++ {
		if err := perunio.Decode(r, (*channel.IDMap)(&(*a)[i])); err != nil {
			return errors.WithMessagef(err, "decoding %d-th id map entry", i)
		}
	}
	return nil
}

// Encode encodes the object to the writer.
func (a indexMapsWithLen) Encode(w io.Writer) (err error) {
	err = perunio.Encode(w, sliceLen(len(a)))
	if err != nil {
		return
	}

	for _, m := range a {
		err = perunio.Encode(w, indexMapWithLen(m))
		if err != nil {
			return
		}
	}
	return
}

// Decode decodes the object from the reader.
func (a *indexMapsWithLen) Decode(r io.Reader) (err error) {
	var l sliceLen
	if err = perunio.Decode(r, &l); err != nil {
		return errors.WithMessage(err, "decoding length")
	}

	*a = make(indexMapsWithLen, l)
	for i := range *a {
		if err = perunio.Decode(r, (*indexMapWithLen)(&(*a)[i])); err != nil {
			return
		}
	}
	return
}

// Encode encodes the object to the writer.
func (a indexMapWithLen) Encode(w io.Writer) (err error) {
	err = perunio.Encode(w, sliceLen(len(a)))
	if err != nil {
		return
	}

	for _, b := range a {
		err = perunio.Encode(w, b)
		if err != nil {
			return
		}
	}
	return
}

// Decode decodes the object from the reader.
func (a *indexMapWithLen) Decode(r io.Reader) (err error) {
	var l sliceLen
	if err = perunio.Decode(r, &l); err != nil {
		return errors.WithMessage(err, "decoding length")
	}

	*a = make(indexMapWithLen, l)
	for i := range *a {
		var b channel.Index
		err = perunio.Decode(r, &b)
		if err != nil {
			return errors.WithMessagef(err, "decoding item %d", i)
		}
		(*a)[i] = b
	}
	return
}
