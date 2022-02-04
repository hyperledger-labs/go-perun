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

package wire_test

import (
	"testing"

	"perun.network/go-perun/wire"
	_ "perun.network/go-perun/wire/perunio/serializer" // wire serialzer init
	peruniotest "perun.network/go-perun/wire/perunio/test"
)

func TestPingMsg(t *testing.T) {
	peruniotest.MsgSerializerTest(t, wire.NewPingMsg())
}

func TestPongMsg(t *testing.T) {
	peruniotest.MsgSerializerTest(t, wire.NewPongMsg())
}

func TestShutdownMsg(t *testing.T) {
	peruniotest.MsgSerializerTest(t, &wire.ShutdownMsg{"m2384ordkln fb30954390582"})
}
