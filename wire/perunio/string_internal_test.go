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
	"encoding/binary"
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	polytest "polycry.pt/poly-go/test"
)

func TestEncodeDecodeString(t *testing.T) {
	assert := assert.New(t)
	rng := polytest.Prng(t)
	uint8buf, uint16buf := make([]byte, math.MaxUint8), make([]byte, math.MaxUint16)
	rng.Read(uint8buf)
	rng.Read(uint16buf)

	t.Run("valid strings", func(t *testing.T) {
		ss := []string{"", "a", "perun", string(uint8buf), string(uint16buf)}

		for _, s := range ss {
			r, w := io.Pipe()
			done := make(chan struct{})
			go func() {
				defer close(done)
				defer w.Close()
				assert.NoError(encodeString(w, s))
			}()

			var d string
			require.NoError(t, decodeString(r, &d))
			r.Close()
			assert.Equal(s, d)
			<-done
		}
	})

	t.Run("too long string", func(t *testing.T) {
		tooLong := string(append(uint16buf, 42)) //nolint: makezero
		var buf bytes.Buffer
		require.Error(t, encodeString(&buf, tooLong))
		assert.Zero(buf.Len(), "nothing should have been written to the stream")
	})

	t.Run("short stream", func(t *testing.T) {
		var buf bytes.Buffer
		err := binary.Write(&buf, byteOrder, uint16(16))
		require.NoError(t, err)
		buf.Write(make([]byte, 8)) // 8 bytes missing

		var d string
		require.Error(t, decodeString(&buf, &d))
		assert.Zero(buf.Len(), "buffer should be exhausted")
	})
}
