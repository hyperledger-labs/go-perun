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
