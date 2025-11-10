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
	"math"

	"github.com/pkg/errors"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire/perunio"
)

type (
	sliceLen          = uint16
	channelIDsWithLen []channel.ID
	indexMapWithLen   []channel.Index
	indexMapsWithLen  [][]channel.Index
)

// Encode encodes the object to the writer.
func (a channelIDsWithLen) Encode(w io.Writer) (err error) {
	l := len(a)
	if l > math.MaxUint16 {
		return errors.New("slice length too long")
	}
	err = perunio.Encode(w, sliceLen(l))
	if err != nil {
		return
	}

	for _, id := range a {
		err = perunio.Encode(w, id)
		if err != nil {
			return
		}
	}
	return
}

// Decode decodes the object from the reader.
func (a *channelIDsWithLen) Decode(r io.Reader) (err error) {
	var l sliceLen
	if err = perunio.Decode(r, &l); err != nil {
		return errors.WithMessage(err, "decoding length")
	}

	*a = make(channelIDsWithLen, l)
	for i := range *a {
		var id channel.ID
		err = perunio.Decode(r, &id)
		if err != nil {
			return errors.WithMessagef(err, "decoding item %d", i)
		}
		(*a)[i] = id
	}
	return
}

// Encode encodes the object to the writer.
func (a indexMapsWithLen) Encode(w io.Writer) (err error) {
	l := len(a)
	if l > math.MaxUint16 {
		return errors.New("slice length too long")
	}
	err = perunio.Encode(w, sliceLen(l))
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
	l := len(a)
	if l > math.MaxUint16 {
		return errors.New("slice length too long")
	}
	err = perunio.Encode(w, sliceLen(l))
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
