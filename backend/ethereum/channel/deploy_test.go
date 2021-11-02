// Copyright 2021 - See NOTICE file for copyright holders.
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

package channel_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func TestDeployReorgResistance(t *testing.T) {
	rng := pkgtest.Prng(t)

	txFinalityDepth := uint64(TxFinalityDepthMin + rng.Intn(TxFinalityDepthMax-TxFinalityDepthMin+1))
	t.Logf("txFinalityDepth = %v", txFinalityDepth)
	s := test.NewSimSetup(t, rng, txFinalityDepth, 0, test.WithCommitTx(false))

	deploy := make(chan error)
	go func() {
		_, err := ethchannel.DeployAdjudicator(context.Background(), *s.CB, s.TxSender.Account)
		deploy <- err
	}()

	time.Sleep(100 * time.Millisecond) // Ensure that deployment transaction is dispatched.
	for i := 0; i < int(txFinalityDepth); i++ {
		select {
		case <-deploy:
			t.Fatal("deploy done before commit")
		case <-time.After(300 * time.Millisecond): // Ensure that last commit is processed.
		}
		s.SimBackend.Commit()
	}

	assert.NoError(t, <-deploy, "deployment")
}
