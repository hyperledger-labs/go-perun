// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/channel"
)

func TestTransferBal(t *testing.T) {
	bals := []channel.Bal{big.NewInt(1000), big.NewInt(500)}
	amount := big.NewInt(42)
	transferBal(bals, 0, amount)
	assert.Equal(t, uint64(958), bals[0].Uint64())
	assert.Equal(t, uint64(542), bals[1].Uint64())
	assert.Equal(t, uint64(42), amount.Uint64())
}
