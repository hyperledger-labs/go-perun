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

package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	perunio "perun.network/go-perun/wire/perunio/serializer"
	"perun.network/go-perun/wire/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestDialer_Dial(t *testing.T) {
	rng := pkgtest.Prng(t)
	ser := perunio.Serializer()
	// Closed dialer must always fail.
	t.Run("closed", func(t *testing.T) {
		var d Dialer
		d.Close()

		conn, err := d.Dial(context.Background(), test.NewRandomAddress(rng), ser)
		assert.Nil(t, conn)
		assert.Error(t, err)
	})

	// Cancelling the context must result in error.
	t.Run("cancel", func(t *testing.T) {
		var d Dialer
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		conn, err := d.Dial(ctx, test.NewRandomAddress(rng), ser)
		assert.Nil(t, conn)
		assert.Error(t, err)
	})
}

func TestDialer_Close(t *testing.T) {
	var d Dialer
	assert.NoError(t, d.Close())
	assert.Error(t, d.Close())
}
