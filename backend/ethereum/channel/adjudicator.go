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

package channel

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	psync "perun.network/go-perun/pkg/sync"
)

// compile time check that we implement the perun adjudicator interface.
var _ channel.Adjudicator = (*Adjudicator)(nil)

// The Adjudicator struct implements the channel.Adjudicator interface
// It provides all functionality to close a channel.
type Adjudicator struct {
	ContractBackend
	contract *adjudicator.Adjudicator
	// The address to which we send all funds.
	Receiver common.Address
	// Structured logger
	log log.Logger
	// Transaction mutex
	mu psync.Mutex
	// txSender is sending the TX.
	txSender accounts.Account
}

// NewAdjudicator creates a new ethereum adjudicator. The receiver is the
// on-chain address that receives withdrawals.
func NewAdjudicator(backend ContractBackend, contract common.Address, receiver common.Address, txSender accounts.Account) *Adjudicator {
	contr, err := adjudicator.NewAdjudicator(contract, backend)
	if err != nil {
		panic("Could not create a new instance of adjudicator")
	}
	return &Adjudicator{
		ContractBackend: backend,
		contract:        contr,
		Receiver:        receiver,
		txSender:        txSender,
		log:             log.WithField("txSender", txSender.Address),
	}
}

func (a *Adjudicator) callRegister(ctx context.Context, req channel.AdjudicatorReq) error {
	return a.call(ctx, req, a.contract.Register)
}

func (a *Adjudicator) callRefute(ctx context.Context, req channel.AdjudicatorReq) error {
	return a.call(ctx, req, a.contract.Refute)
}

func (a *Adjudicator) callConclude(ctx context.Context, req channel.AdjudicatorReq) error {
	// Wrapped call to Conclude, ignoring sig
	conclude := func(
		opts *bind.TransactOpts,
		params adjudicator.ChannelParams,
		state adjudicator.ChannelState,
		_ [][]byte,
	) (*types.Transaction, error) {
		return a.contract.Conclude(opts, params, state, []adjudicator.ChannelState{})
	}
	return a.call(ctx, req, conclude)
}

func (a *Adjudicator) callConcludeFinal(ctx context.Context, req channel.AdjudicatorReq) error {
	return a.call(ctx, req, a.contract.ConcludeFinal)
}

type adjFunc = func(
	opts *bind.TransactOpts,
	params adjudicator.ChannelParams,
	state adjudicator.ChannelState,
	sigs [][]byte,
) (*types.Transaction, error)

// call calls the given contract function `fn` with the data from `req`.
// `fn` should be a method of `a.contract`, like `a.contract.Register`.
func (a *Adjudicator) call(ctx context.Context, req channel.AdjudicatorReq, fn adjFunc) error {
	ethParams := ToEthParams(req.Params)
	ethState := ToEthState(req.Tx.State)
	tx, err := func() (*types.Transaction, error) {
		if !a.mu.TryLockCtx(ctx) {
			return nil, errors.Wrap(ctx.Err(), "context canceled while acquiring tx lock")
		}
		defer a.mu.Unlock()

		trans, err := a.NewTransactor(ctx, GasLimit, a.txSender)
		if err != nil {
			return nil, errors.WithMessage(err, "creating transactor")
		}
		tx, err := fn(trans, ethParams, ethState, req.Tx.Sigs)
		if err != nil {
			return nil, errors.Wrap(err, "calling adjudicator function")
		}
		log.Debugf("Sent transaction %v", tx.Hash().Hex())
		return tx, nil
	}()
	if err != nil {
		return err
	}

	return errors.WithMessage(a.ConfirmTransaction(ctx, tx, a.txSender), "mining transaction")
}

// Progress will call progress on the Adjudicator contract in the future. Right
// now, it panics.
func (a *Adjudicator) Progress(context.Context, channel.ProgressReq) error {
	a.log.Panic("Progressed not implemented yet")
	return nil
}

// ValidateAdjudicator checks if the bytecode at given address is correct.
// Returns a ContractBytecodeError if the bytecode at given address is invalid.
// This error can be checked with function IsErrInvalidContractCode.
func ValidateAdjudicator(ctx context.Context,
	backend bind.ContractCaller, adjudicatorAddr common.Address) error {
	return validateContract(ctx, backend, adjudicatorAddr, adjudicator.AdjudicatorBinRuntime)
}
