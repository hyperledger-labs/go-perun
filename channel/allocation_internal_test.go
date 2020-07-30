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

package channel

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

// simple summer for testing.
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
