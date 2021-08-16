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
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/sync"
)

const (
	// GasLimit is the max amount of gas we want to send per transaction.
	GasLimit = 500000

	// GasPrice is the gas price that is used for simulated transactions.
	// This value is set to `maxFeePerGas` from go-ethereum to prevent
	// "max fee per gas less than block base fee" errors.
	GasPrice = 875000000
)

// SimulatedBackend provides a simulated ethereum blockchain for tests.
type SimulatedBackend struct {
	backends.SimulatedBackend
	sbMtx sync.Mutex // protects SimulatedBackend

	faucetKey  *ecdsa.PrivateKey
	faucetAddr common.Address
	clockMu    sync.Mutex    // Mutex for clock adjustments. Locked by SimTimeouts.
	mining     chan struct{} // Used for auto-mining blocks.
}

// Reorder can be used to insert, reorder and exclude transactions in
// combination with `Reorg`.
type Reorder func([]types.Transactions) []types.Transactions

// NewSimulatedBackend creates a new Simulated Backend.
func NewSimulatedBackend() *SimulatedBackend {
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	faucetAddr := crypto.PubkeyToAddress(sk.PublicKey)
	addr := map[common.Address]core.GenesisAccount{
		common.BytesToAddress([]byte{1}): {Balance: big.NewInt(1)}, // ECRecover
		common.BytesToAddress([]byte{2}): {Balance: big.NewInt(1)}, // SHA256
		common.BytesToAddress([]byte{3}): {Balance: big.NewInt(1)}, // RIPEMD
		common.BytesToAddress([]byte{4}): {Balance: big.NewInt(1)}, // Identity
		common.BytesToAddress([]byte{5}): {Balance: big.NewInt(1)}, // ModExp
		common.BytesToAddress([]byte{6}): {Balance: big.NewInt(1)}, // ECAdd
		common.BytesToAddress([]byte{7}): {Balance: big.NewInt(1)}, // ECScalarMul
		common.BytesToAddress([]byte{8}): {Balance: big.NewInt(1)}, // ECPairing
		faucetAddr:                       {Balance: new(big.Int).Sub(channel.MaxBalance, big.NewInt(9))},
	}
	alloc := core.GenesisAlloc(addr)
	return &SimulatedBackend{
		SimulatedBackend: *backends.NewSimulatedBackend(alloc, 8000000),
		faucetKey:        sk,
		faucetAddr:       faucetAddr,
	}
}

// SuggestGasPrice always returns `GasPrice`.
func (*SimulatedBackend) SuggestGasPrice(context.Context) (*big.Int, error) {
	return big.NewInt(GasPrice), nil
}

// SendTransaction executes a transaction.
func (s *SimulatedBackend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	if err := s.SimulatedBackend.SendTransaction(ctx, tx); err != nil {
		return errors.WithStack(err)
	}
	s.Commit()
	return nil
}

// FundAddress funds a given address with `test.MaxBalance` eth from a faucet.
func (s *SimulatedBackend) FundAddress(ctx context.Context, addr common.Address) {
	nonce, err := s.PendingNonceAt(context.Background(), s.faucetAddr)
	if err != nil {
		panic(err)
	}
	tx := types.NewTransaction(nonce, addr, test.MaxBalance, GasLimit, big.NewInt(GasPrice), nil)
	signer := types.NewEIP155Signer(big.NewInt(1337))
	signedTX, err := types.SignTx(tx, signer, s.faucetKey)
	if err != nil {
		panic(err)
	}
	if err := s.SendTransaction(ctx, signedTX); err != nil {
		panic(err)
	}
	bind.WaitMined(context.Background(), s, signedTX)
}

// StartMining makes the simulated blockchain auto-mine blocks with the given
// interval. Must be stopped with `StopMining`.
// The block time of generated blocks will always increase by 10 seconds.
func (s *SimulatedBackend) StartMining(interval time.Duration) {
	if interval == 0 {
		panic("blockTime can not be zero")
	}

	s.mining = make(chan struct{})
	go func() {
		log.Trace("Started mining")
		defer log.Trace("Stopped mining")

		for {
			s.Commit()
			log.Trace("Mined simulated block")

			select {
			case <-time.After(interval):
			case <-s.mining: // stopped
				return
			}
		}
	}()
}

// StopMining stops the auto-mining of the simulated blockchain.
// Must be called exactly once to free resources iff `StartMining` was called.
func (s *SimulatedBackend) StopMining() {
	close(s.mining)
}

// Reorg applies a chain reorg.
// `depth` is the number of blocks to be removed.
// `reorder` is a function that gets as input the removed blocks and outputs a list of blocks that are to be added after the removal.
// It is required that the number of added blocks is greater than `depth` for a reorg to be accepted.
// The nonce prevents transactions of the same account from being re-ordered. Trying to do this will panic.
func (s *SimulatedBackend) Reorg(ctx context.Context, depth uint64, reorder Reorder) error {
	// Lock
	if !s.sbMtx.TryLockCtx(ctx) {
		return errors.Errorf("locking mutex: %v", ctx.Err())
	}
	defer s.sbMtx.Unlock()

	// parent at current - depth.
	parentN := new(big.Int).Sub(s.Blockchain().CurrentBlock().Number(), big.NewInt(int64(depth)))
	parent, err := s.BlockByNumber(ctx, parentN)
	if err != nil {
		return errors.Wrap(err, "retrieving reorg parent")
	}

	// Collect orphaned blocks.
	blocks := make([]types.Transactions, depth)
	for i := uint64(0); i < depth; i++ {
		blockN := new(big.Int).Add(parentN, big.NewInt(int64(i+1)))
		block, err := s.BlockByNumber(ctx, blockN)
		if err != nil {
			return errors.Wrap(err, "retrieving block")
		}
		// Add the TXs from block parent + 1 + i.
		blocks[i] = block.Transactions()
	}

	// Modify the blocks with the reorder callback.
	newBlocks := reorder(blocks)
	if uint64(len(newBlocks)) <= depth {
		return fmt.Errorf("number of blocks added %d must be greater than number of blocks removed %d", len(newBlocks), depth)
	}

	// Reset the chain to the parent block.
	if err := s.Fork(ctx, parent.Hash()); err != nil {
		return errors.Wrap(err, "forking")
	}

	// Add modified blocks.
	for _, txs := range newBlocks {
		for _, tx := range txs {
			if err := s.SimulatedBackend.SendTransaction(ctx, tx); err != nil {
				return errors.Wrap(err, "re-sending transaction")
			}
		}
		s.Commit()
	}
	return nil
}
