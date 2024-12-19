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

package test

import (
	"testing"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wire"
	pkgtest "polycry.pt/poly-go/test"
)

// ChannelSyncMsgSerializationTest runs serialization tests on channel sync messages.
func ChannelSyncMsgSerializationTest(t *testing.T, serializerTest func(t *testing.T, msg wire.Msg)) {
	t.Helper()
	rng := pkgtest.Prng(t)
	for i := 0; i < 4; i++ {
		state := test.NewRandomState(rng)
		m := &client.ChannelSyncMsg{
			Phase: channel.Phase(rng.Intn(channel.LastPhase)), //nolint:gosec
			CurrentTX: channel.Transaction{
				State: state,
				Sigs:  newRandomSigs(rng, state.NumParts()),
			},
		}
		serializerTest(t, m)
	}
}
