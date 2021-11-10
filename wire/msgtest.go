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

package wire

import (
	"io"
	"testing"

	"polycry.pt/poly-go/io/test"
)

type serializerMsg struct {
	Msg Msg
}

func (msg *serializerMsg) Encode(writer io.Writer) error {
	return Encode(msg.Msg, writer)
}

func (msg *serializerMsg) Decode(reader io.Reader) (err error) {
	msg.Msg, err = Decode(reader)
	return err
}

// TestMsg performs generic tests on a wire.Msg object.
func TestMsg(t *testing.T, msg Msg) {
	test.GenericSerializerTest(t, &serializerMsg{msg})
}
