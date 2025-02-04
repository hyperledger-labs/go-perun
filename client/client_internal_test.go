// Copyright 2025 - See NOTICE file for copyright holders.
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

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"

	channeltest "perun.network/go-perun/channel/test"
	wiretest "perun.network/go-perun/wire/test"
	"polycry.pt/poly-go/test"
)

func TestClient_Channel(t *testing.T) {
	rng := test.Prng(t)
	// dummy client that only has an id and a registry
	c := &Client{
		address:  wiretest.NewRandomAddressesMap(rng, 1)[0],
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
