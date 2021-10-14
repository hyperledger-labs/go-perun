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

package test

import (
	"io"
	"testing"

	"perun.network/go-perun/wire"
)

type serializerMsg struct {
	Msg wire.Msg
}

func (msg *serializerMsg) Encode(writer io.Writer) error {
	return wire.Encode(msg.Msg, writer)
}

func (msg *serializerMsg) Decode(reader io.Reader) (err error) {
	msg.Msg, err = wire.Decode(reader)
	return err
}

// TestMsgSerializer performs generic serializer tests on a wire.Msg object.
func TestMsgSerializer(t *testing.T, msg wire.Msg) {
	GenericSerializerTest(t, &serializerMsg{msg})
}
