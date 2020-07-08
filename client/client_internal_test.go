// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package client

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	channeltest "perun.network/go-perun/channel/test"
	wallettest "perun.network/go-perun/wallet/test"
)

func TestClient_Channel(t *testing.T) {
	rng := rand.New(rand.NewSource(0xdeadbeef))
	id := wallettest.NewRandomAccount(rng)
	// dummy client that only has an id and a registry
	c := &Client{
		id:       id,
		channels: makeChanRegistry(),
	}

	cID := channeltest.NewRandomChannelID(rng)

	t.Run("unknown", func(t *testing.T) {
		ch, err := c.Channel(cID)
		assert.Nil(t, ch)
		assert.Error(t, err)
	})

	t.Run("known", func(t *testing.T) {
		ch1 := testCh()
		c.channels.Put(cID, ch1)

		ch2, err := c.Channel(cID)
		assert.Same(t, ch2, ch1)
		assert.NoError(t, err)
	})
}
