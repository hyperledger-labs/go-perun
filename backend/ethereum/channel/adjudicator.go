// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	psync "perun.network/go-perun/pkg/sync"
)

// compile time check that we implement the perun adjudicator interface
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
}

// NewAdjudicator creates a new ethereum adjudicator. The receiver is the
// on-chain address that receives withdrawals.
func NewAdjudicator(backend ContractBackend, contract common.Address, receiver common.Address) *Adjudicator {
	contr, err := adjudicator.NewAdjudicator(contract, backend)
	if err != nil {
		panic("Could not create a new instance of adjudicator")
	}
	return &Adjudicator{
		ContractBackend: backend,
		contract:        contr,
		Receiver:        receiver,
		log:             log.WithField("account", backend.account.Address),
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
		return a.contract.Conclude(opts, params, state)
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
	ethParams := channelParamsToEthParams(req.Params)
	ethState := channelStateToEthState(req.Tx.State)
	tx, err := func() (*types.Transaction, error) {
		if !a.mu.TryLockCtx(ctx) {
			return nil, errors.Wrap(ctx.Err(), "context canceled while acquiring tx lock")
		}
		defer a.mu.Unlock()

		trans, err := a.NewTransactor(ctx, big.NewInt(0), GasLimit)
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

	return errors.WithMessage(confirmTransaction(ctx, a.ContractBackend, tx), "mining transaction")
}
