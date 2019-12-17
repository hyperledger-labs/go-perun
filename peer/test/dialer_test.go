// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/wallet/test"
)

func TestDialer_Dial(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDDDface))
	// Closed dialer must always fail.
	t.Run("closed", func(t *testing.T) {
		var d Dialer
		d.Close()

		conn, err := d.Dial(context.Background(), test.NewRandomAddress(rng))
		assert.Nil(t, conn)
		assert.Error(t, err)
	})

	// Cancelling the context must result in error.
	t.Run("cancel", func(t *testing.T) {
		var d Dialer
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		conn, err := d.Dial(ctx, test.NewRandomAddress(rng))
		assert.Nil(t, conn)
		assert.Error(t, err)
	})
}

func TestDialer_Close(t *testing.T) {
	var d Dialer
	assert.NoError(t, d.Close())
	assert.Error(t, d.Close())
}
