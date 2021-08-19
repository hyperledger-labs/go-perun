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

package client_test

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/wallet"
	chtest "perun.network/go-perun/channel/test"
	perunclient "perun.network/go-perun/client"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/log"
	pkgtest "perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wire"
)

var defaultTimeout = 5 * time.Second

func TestHappyAliceBob(t *testing.T) {
	log.Info("Starting happy test")
	rng := pkgtest.Prng(t)

	const A, B = 0, 1 // Indices of Alice and Bob
	var (
		name  = [2]string{"Alice", "Bob"}
		setup [2]clienttest.RoleSetup
		role  [2]clienttest.Executer
	)

	s := test.NewSetup(t, rng, 2, blockInterval)
	setup = makeRoleSetups(s, name)

	role[A] = clienttest.NewAlice(setup[A], t)
	role[B] = clienttest.NewBob(setup[B], t)
	// enable stages synchronization
	stages := role[A].EnableStages()
	role[B].SetStages(stages)

	execConfig := &clienttest.AliceBobExecConfig{
		BaseExecConfig: clienttest.MakeBaseExecConfig(
			[2]wire.Address{setup[A].Identity.Address(), setup[B].Identity.Address()},
			(*wallet.Address)(&s.Asset),
			[2]*big.Int{big.NewInt(100), big.NewInt(100)},
			perunclient.WithApp(chtest.NewRandomAppAndData(rng)),
		),
		NumPayments: [2]int{2, 2},
		TxAmounts:   [2]*big.Int{big.NewInt(5), big.NewInt(3)},
	}

	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func(i int) {
			defer wg.Done()
			log.Infof("Starting %s.Execute", name[i])
			role[i].Execute(execConfig)
		}(i)
	}

	wg.Wait()

	// Assert correct final balances
	aliceToBob := big.NewInt(int64(execConfig.NumPayments[A])*execConfig.TxAmounts[A].Int64() -
		int64(execConfig.NumPayments[B])*execConfig.TxAmounts[B].Int64())
	finalBalAlice := new(big.Int).Sub(execConfig.InitBals()[A], aliceToBob)
	finalBalBob := new(big.Int).Add(execConfig.InitBals()[B], aliceToBob)
	// reset context timeout
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	assertBal := func(addr *wallet.Address, bal *big.Int) {
		b, err := s.SimBackend.BalanceAt(ctx, common.Address(*addr), nil)
		require.NoError(t, err)
		assert.Zero(t, bal.Cmp(b), "ETH balance mismatch")
	}

	assertBal(s.Recvs[A], finalBalAlice)
	assertBal(s.Recvs[B], finalBalBob)

	log.Info("Happy test done")
}
