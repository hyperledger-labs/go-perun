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

package perunio

import (
	"bytes"

	"github.com/pkg/errors"
)

// EqualEncoding returns whether the two Encoders `a` and `b` encode to the same byteslice
// or an error when the encoding failed.
func EqualEncoding(a, b Encoder) (bool, error) {
	buffA := new(bytes.Buffer)
	buffB := new(bytes.Buffer)

	// golang does not have a XOR
	if (a == nil) != (b == nil) {
		return false, errors.New("only one argument was nil")
	}
	// just using a == b would be too easy here since go panics
	if (a == nil) && (b == nil) {
		return true, nil
	}

	err := a.Encode(buffA)
	if err != nil {
		return false, errors.Wrap(err, "EqualEncoding encode error")
	}

	err = b.Encode(buffB)
	if err != nil {
		return false, errors.Wrap(err, "EqualEncoding encode error")
	}

	return bytes.Equal(buffA.Bytes(), buffB.Bytes()), nil
}
