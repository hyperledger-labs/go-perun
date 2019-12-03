// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package client_test

import (
	"math/rand"
	"testing"

	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire/msg"
)

func TestChannelUpdateSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(0xdeadbeef))
	for i := 0; i < 4; i++ {
		params := test.NewRandomParams(rng, test.NewRandomApp(rng).Def())
		sig, _ := wallet.DecodeSig(rng)
		m := &client.ChannelUpdate{
			State:    test.NewRandomState(rng, params),
			ActorIdx: uint16(rng.Int31n(int32(len(params.Parts)))),
			Sig:      sig,
		}
		msg.TestMsg(t, m)
	}
}

func TestChannelUpdateAccSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(0xc0ffeefee))
	for i := 0; i < 4; i++ {
		sig, _ := wallet.DecodeSig(rng)
		m := &client.ChannelUpdateAcc{
			ChannelID: test.NewRandomChannelID(rng),
			Version:   uint64(rng.Int63()),
			Sig:       sig,
		}
		msg.TestMsg(t, m)
	}
}

func TestChannelUpdateRejSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(0xdeadbeef))
	for i := 0; i < 4; i++ {
		params := test.NewRandomParams(rng, test.NewRandomApp(rng).Def())
		sig, _ := wallet.DecodeSig(rng)
		m := &client.ChannelUpdateRej{
			Reason:   newRandomString(rng, 16, 16),
			Alt:      test.NewRandomState(rng, params),
			ActorIdx: uint16(rng.Int31n(int32(len(params.Parts)))),
			Sig:      sig,
		}
		msg.TestMsg(t, m)
	}
}
