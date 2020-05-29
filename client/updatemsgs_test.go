// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package client

import (
	"math/rand"
	"testing"

	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire/msg"
)

func TestChannelUpdateSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(0xdeadbeef))
	for i := 0; i < 4; i++ {
		params, state := test.NewRandomParamsAndState(rng)
		sig := newRandomSig(rng)
		m := &msgChannelUpdate{
			ChannelUpdate: ChannelUpdate{
				State:    state,
				ActorIdx: uint16(rng.Int31n(int32(len(params.Parts)))),
			},
			Sig: sig,
		}
		msg.TestMsg(t, m)
	}
}

func TestChannelUpdateAccSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(0xc0ffeefee))
	for i := 0; i < 4; i++ {
		sig := newRandomSig(rng)
		m := &msgChannelUpdateAcc{
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
		m := &msgChannelUpdateRej{
			ChannelID: test.NewRandomChannelID(rng),
			Version:   uint64(rng.Int63()),
			Reason:    newRandomString(rng, 16, 16),
		}
		msg.TestMsg(t, m)
	}
}

// newRandomSig generates a random account and then returns the signature on
// some random data.
func newRandomSig(rng *rand.Rand) wallet.Sig {
	acc := wallettest.NewRandomAccount(rng)
	data := make([]byte, 8)
	rng.Read(data)
	sig, err := acc.SignData(data)
	if err != nil {
		panic("signing error")
	}
	return sig
}

// newRandomstring returns a random string of length between minLen and
// minLen+maxLenDiff.
func newRandomString(rng *rand.Rand, minLen, maxLenDiff int) string {
	r := make([]byte, minLen+rng.Intn(maxLenDiff))
	rng.Read(r)
	return string(r)
}
