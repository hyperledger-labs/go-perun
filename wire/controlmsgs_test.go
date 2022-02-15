// Copyright 2022 - See NOTICE file for copyright holders.
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

	_ "perun.network/go-perun/wire/perunio/serializer" // wire serialzer init
	peruniotest "perun.network/go-perun/wire/perunio/test"
	wiretest "perun.network/go-perun/wire/test"
)

func TestControlMsgs(t *testing.T) {
	wiretest.ControlMsgsSerializationTest(t, peruniotest.MsgSerializerTest)
}
