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
	protoMsg := &PingMsg{
		Created: msg.Created.UnixNano(),
	}
	return &Envelope_PingMsg{protoMsg}
}

func fromPongMsg(msg *wire.PongMsg) *Envelope_PongMsg {
	protoMsg := &PongMsg{
		Created: msg.Created.UnixNano(),
	}
	return &Envelope_PongMsg{protoMsg}
}

func fromShutdownMsg(msg *wire.ShutdownMsg) *Envelope_ShutdownMsg {
	protoMsg := &ShutdownMsg{
		Reason: msg.Reason,
	}
	return &Envelope_ShutdownMsg{protoMsg}
}

func fromAuthResponseMsg(msg *wire.AuthResponseMsg) *Envelope_AuthResponseMsg {
	protoMsg := &AuthResponseMsg{}
	protoMsg.Signature = msg.Signature
	protoMsg.SignatureSize = msg.SignatureSize
	return &Envelope_AuthResponseMsg{protoMsg}
}

//nolint:forbidigo
func toPingMsg(protoMsg *Envelope_PingMsg) *wire.PingMsg {
	msg := &wire.PingMsg{}
	msg.Created = time.Unix(0, protoMsg.PingMsg.GetCreated())
	return msg
}

//nolint:forbidigo
func toPongMsg(protoEnvMsg *Envelope_PongMsg) *wire.PongMsg {
	msg := &wire.PongMsg{}
	msg.Created = time.Unix(0, protoEnvMsg.PongMsg.GetCreated())
	return msg
}

//nolint:forbidigo
func toShutdownMsg(protoEnvMsg *Envelope_ShutdownMsg) *wire.ShutdownMsg {
	msg := &wire.ShutdownMsg{}
	msg.Reason = protoEnvMsg.ShutdownMsg.GetReason()
	return msg
}

//nolint:forbidigo
func toAuthResponseMsg(protoEnvMsg *Envelope_AuthResponseMsg) *wire.AuthResponseMsg {
	msg := &wire.AuthResponseMsg{}
	msg.SignatureSize = protoEnvMsg.AuthResponseMsg.GetSignatureSize()
	msg.Signature = protoEnvMsg.AuthResponseMsg.GetSignature()
	if msg.Signature == nil {
		msg.Signature = make([]byte, msg.SignatureSize)
	}
	return msg
}
