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

package protobuf_test

import (
	"testing"

	_ "perun.network/go-perun/backend/sim/channel"
	_ "perun.network/go-perun/backend/sim/wallet"
	clienttest "perun.network/go-perun/client/test"
	protobuftest "perun.network/go-perun/wire/protobuf/test"
	wiretest "perun.network/go-perun/wire/test"
)

func TestControlMsgsSerialization(t *testing.T) {
	wiretest.ControlMsgsSerializationTest(t, protobuftest.MsgSerializerTest)
}

func TestAuthResponseMsgSerialization(t *testing.T) {
	wiretest.AuthMsgsSerializationTest(t, protobuftest.MsgSerializerTest)
}

func TestProposalMsgsSerialization(t *testing.T) {
	clienttest.ProposalMsgsSerializationTest(t, protobuftest.MsgSerializerTest)
}

func TestUpdateMsgsSerialization(t *testing.T) {
	clienttest.UpdateMsgsSerializationTest(t, protobuftest.MsgSerializerTest)
}

func TestChannelSyncMsgSerialization(t *testing.T) {
	clienttest.ChannelSyncMsgSerializationTest(t, protobuftest.MsgSerializerTest)
}
