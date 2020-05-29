// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package channel

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

// simple summer for testing
type balsum struct {
	b []Bal
}

func (b balsum) Sum() []Bal {
	return b.b
}

func TestEqualBalance(t *testing.T) {
	empty := balsum{make([]Bal, 0)}
	one1 := balsum{[]Bal{big.NewInt(1)}}
	one2 := balsum{[]Bal{big.NewInt(2)}}
	two12 := balsum{[]Bal{big.NewInt(1), big.NewInt(2)}}
	two48 := balsum{[]Bal{big.NewInt(4), big.NewInt(8)}}

	assert := assert.New(t)

	_, err := equalSum(empty, one1)
	assert.NotNil(err)

	eq, err := equalSum(empty, empty)
	assert.Nil(err)
	assert.True(eq)

	eq, err = equalSum(one1, one1)
	assert.Nil(err)
	assert.True(eq)

	eq, err = equalSum(one1, one2)
	assert.Nil(err)
	assert.False(eq)

	_, err = equalSum(one1, two12)
	assert.NotNil(err)

	eq, err = equalSum(two12, two12)
	assert.Nil(err)
	assert.True(eq)

	eq, err = equalSum(two12, two48)
	assert.Nil(err)
	assert.False(eq)
}
