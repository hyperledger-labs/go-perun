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
	"bytes"
	"encoding"

	"github.com/pkg/errors"
)

// EqualBinary returns whether the binary representation of the two values `a` and `b` are equal.
// It returns an error when marshalling fails.
func EqualBinary(a, b encoding.BinaryMarshaler) (bool, error) {
	// golang does not have a XOR
	if (a == nil) != (b == nil) {
		return false, errors.New("only one argument was nil")
	}
	// just using a == b would be too easy here since go panics
	if (a == nil) && (b == nil) {
		return true, nil
	}

	binaryA, err := a.MarshalBinary()
	if err != nil {
		return false, errors.Wrap(err, "EqualBinary: marshaling a")
	}
	binaryB, err := b.MarshalBinary()
	if err != nil {
		return false, errors.Wrap(err, "EqualBinary: marshaling b")
	}

	return bytes.Equal(binaryA, binaryB), nil
}
