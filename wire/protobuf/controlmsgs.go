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

package protobuf

import (
	"time"

	"perun.network/go-perun/wire"
)

func fromPingMsg(msg *wire.PingMsg) *Envelope_PingMsg {
	protoMsg := &PingMsg{}
	protoMsg.Created = msg.Created.UnixNano()
	return &Envelope_PingMsg{protoMsg}
}

func fromPongMsg(msg *wire.PongMsg) *Envelope_PongMsg {
	protoMsg := &PongMsg{}
	protoMsg.Created = msg.Created.UnixNano()
	return &Envelope_PongMsg{protoMsg}
}

func fromShutdownMsg(msg *wire.ShutdownMsg) *Envelope_ShutdownMsg {
	protoMsg := &ShutdownMsg{}
	protoMsg.Reason = msg.Reason
	return &Envelope_ShutdownMsg{protoMsg}
}

func toPingMsg(protoMsg *Envelope_PingMsg) (msg *wire.PingMsg) {
	msg = &wire.PingMsg{}
	msg.Created = time.Unix(0, protoMsg.PingMsg.Created)
	return msg
}

func toPongMsg(protoEnvMsg *Envelope_PongMsg) (msg *wire.PongMsg) {
	msg = &wire.PongMsg{}
	msg.Created = time.Unix(0, protoEnvMsg.PongMsg.Created)
	return msg
}

func toShutdownMsg(protoEnvMsg *Envelope_ShutdownMsg) (msg *wire.ShutdownMsg) {
	msg = &wire.ShutdownMsg{}
	msg.Reason = protoEnvMsg.ShutdownMsg.Reason
	return msg
}
