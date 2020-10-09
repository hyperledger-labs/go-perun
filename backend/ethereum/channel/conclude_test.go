// Copyright 2020 - See NOTICE file for copyright holders.
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

	"github.com/stretchr/testify/require"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func TestAdjudicator_ConcludeFinal(t *testing.T) {
	t.Run("ConcludeFinal 1 party", func(t *testing.T) { testConcludeFinal(t, 1) })
	t.Run("ConcludeFinal 2 party", func(t *testing.T) { testConcludeFinal(t, 2) })
	t.Run("ConcludeFinal 5 party", func(t *testing.T) { testConcludeFinal(t, 5) })
	t.Run("ConcludeFinal 10 party", func(t *testing.T) { testConcludeFinal(t, 10) })
}

func testConcludeFinal(t *testing.T, numParts int) {
	rng := pkgtest.Prng(t)
	// create test setup
	s := test.NewSetup(t, rng, numParts)
	// create valid state and params
	params, state := channeltest.NewRandomParamsAndState(rng, channeltest.WithParts(s.Parts...), channeltest.WithAssets((*ethchannel.Asset)(&s.Asset)), channeltest.WithIsFinal(true))
	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer funCancel()
	// fund the contract
	ct := pkgtest.NewConcurrent(t)
	for i, funder := range s.Funders {
		i, funder := i, funder
		go ct.StageN("funding loop", numParts, func(rt pkgtest.ConcT) {
			req := channel.FundingReq{
				Params: params,
				State:  state,
				Idx:    channel.Index(i),
			}
			require.NoError(rt, funder.Fund(fundingCtx, req), "funding should succeed")
		})
	}
	ct.Wait("funding loop")
	tx := signState(t, s.Accs, params, state)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	ct = pkgtest.NewConcurrent(t)
	initiator := int(rng.Int31n(int32(numParts))) // pick a random initiator
	for i := 0; i < numParts; i++ {
		i := i
		go ct.StageN("register", numParts, func(t pkgtest.ConcT) {
			req := channel.AdjudicatorReq{
				Params:    params,
				Acc:       s.Accs[i],
				Idx:       channel.Index(i),
				Tx:        tx,
				Secondary: (i != initiator),
			}
			diff, err := test.NonceDiff(s.Accs[i].Address(), s.Adjs[i], func() error {
				_, err := s.Adjs[i].Register(ctx, req)
				return err
			})
			require.NoError(t, err, "Withdrawing should succeed")
			if !req.Secondary {
				// The Initiator must send a TX.
				require.Equal(t, diff, 1)
			} else {
				// Everyone else must NOT send a TX.
				require.Equal(t, diff, 0)
			}
		})
	}
	ct.Wait("register")
}
