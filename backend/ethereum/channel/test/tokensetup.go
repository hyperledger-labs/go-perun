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

package test

import (
	"context"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/bindings/peruntoken"
	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/wallet/keystore"
	channeltest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/log"
	wallettest "perun.network/go-perun/wallet/test"
)

type TokenSetup struct {
	SB         *SimulatedBackend
	CB         ethchannel.ContractBackend
	Token      *peruntoken.ERC20
	R          *require.Assertions
	T          *testing.T
	Acc1, Acc2 *accounts.Account

	subApproval, subTransfer event.Subscription
	SinkApproval             chan *peruntoken.ERC20Approval
	SinkTransfer             chan *peruntoken.ERC20Transfer
}

const eventTimeout = 100 * time.Millisecond

func NewTokenSetup(ctx context.Context, t *testing.T, rng *rand.Rand) *TokenSetup {
	// Simulated chain setup.
	sb := NewSimulatedBackend()
	ksWallet := wallettest.RandomWallet().(*keystore.Wallet)
	acc1 := &ksWallet.NewRandomAccount(rng).(*keystore.Account).Account
	sb.FundAddress(ctx, acc1.Address)
	acc2 := &ksWallet.NewRandomAccount(rng).(*keystore.Account).Account
	sb.FundAddress(ctx, acc2.Address)
	cb := ethchannel.NewContractBackend(sb, keystore.NewTransactor(*ksWallet, types.NewEIP155Signer(big.NewInt(1337))))

	// Setup Perun Token.
	tokenAddr, err := ethchannel.DeployPerunToken(ctx, cb, *acc1, []common.Address{acc1.Address}, channeltest.MaxBalance)
	require.NoError(t, err)
	token, err := peruntoken.NewERC20(tokenAddr, cb)
	require.NoError(t, err)

	// Approval sub.
	sinkApproval := make(chan *peruntoken.ERC20Approval, 10)
	subApproval, err := token.WatchApproval(&bind.WatchOpts{Context: ctx}, sinkApproval, nil, nil)
	require.NoError(t, err)
	// Transfer sub.
	sinkTransfer := make(chan *peruntoken.ERC20Transfer, 10)
	subTransfer, err := token.WatchTransfer(&bind.WatchOpts{Context: ctx}, sinkTransfer, nil, nil)
	require.NoError(t, err)

	return &TokenSetup{
		SB:           sb,
		CB:           cb,
		Token:        token,
		R:            require.New(t),
		T:            t,
		Acc1:         acc1,
		Acc2:         acc2,
		SinkApproval: sinkApproval,
		SinkTransfer: sinkTransfer,
		subApproval:  subApproval,
		subTransfer:  subTransfer,
	}
}

const (
	txGasLimit = 100000
)

func (s *TokenSetup) IncAllowance(ctx context.Context) *types.Transaction {
	opts, err := s.CB.NewTransactor(ctx, txGasLimit, *s.Acc1)
	s.R.NoError(err)
	tx, err := s.Token.IncreaseAllowance(opts, s.Acc2.Address, big.NewInt(1))
	s.R.NoError(err)
	return tx
}

func (s *TokenSetup) Transfer(ctx context.Context) *types.Transaction {
	opts, err := s.CB.NewTransactor(ctx, txGasLimit, *s.Acc2)
	s.R.NoError(err)
	tx, err := s.Token.TransferFrom(opts, s.Acc1.Address, s.Acc2.Address, big.NewInt(1))
	s.R.NoError(err)
	return tx
}

func (s *TokenSetup) ConfirmTx(tx *types.Transaction, confirm bool) {
	ctx, cancel := context.WithTimeout(context.Background(), eventTimeout)
	defer cancel()
	rec, err := s.CB.ConfirmTransaction(ctx, tx, *s.Acc1)

	if rec != nil {
		log.Infof("Num: %v, BlockHash: %v", rec.BlockNumber, rec.BlockHash.Hex())
	}
	if confirm {
		s.R.NoError(err)
	} else {
		s.R.Error(err)
	}
}

func (s *TokenSetup) AllowanceEvent(v uint64, included bool) {
	var e *peruntoken.ERC20Approval

	select {
	case e = <-s.SinkApproval:
	case <-time.After(eventTimeout):
		s.T.FailNow()
	}

	s.R.NotNil(e)
	s.R.Equal(s.Acc1.Address, e.Owner)
	s.R.Equal(s.Acc2.Address, e.Spender)
	s.R.Equal(v, e.Value.Uint64())
	s.R.Equal(!included, e.Raw.Removed)
}

func (s *TokenSetup) TransferEvent(included bool) {
	var e *peruntoken.ERC20Transfer

	select {
	case e = <-s.SinkTransfer:
	case <-time.After(eventTimeout):
		s.T.FailNow()
	}

	s.R.NotNil(e)
	s.R.Equal(s.Acc1.Address, e.From)
	s.R.Equal(s.Acc2.Address, e.To)
	s.R.Equal(big.NewInt(1), e.Value)
	s.R.Equal(!included, e.Raw.Removed)
}

func (s *TokenSetup) NoMoreEvents() {
	select {
	case e := <-s.SinkApproval:
		s.R.FailNow("Expected no event but got: ", e)
	case e := <-s.SinkTransfer:
		s.R.FailNow("Expected no event but got: ", e)
	case <-time.After(eventTimeout):
	}
}

func (s *TokenSetup) Close() {
	s.subApproval.Unsubscribe()
	s.subTransfer.Unsubscribe()
	close(s.SinkApproval)
	close(s.SinkTransfer)
}
