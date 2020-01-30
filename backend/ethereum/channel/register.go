// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import (
	"context"
	"math/big"

	"github.com/pkg/errors"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
)

func (a *Adjudicator) register(ctx context.Context, req channel.AdjudicatorReq) error {
	ethParams := channelParamsToEthParams(req.Params)
	ethState := channelStateToEthState(req.Tx.State)
	a.mu.Lock()
	defer a.mu.Unlock()
	trans, err := a.newTransactor(ctx, big.NewInt(0), GasLimit)
	if err != nil {
		return errors.WithMessage(err, "creating transactor")
	}
	tx, err := a.contract.Register(trans, ethParams, ethState, req.Tx.Sigs)
	if err != nil {
		return errors.Wrap(err, "calling concludeFinal")
	}
	log.Debugf("Sending transaction to the blockchain with txHash: %v", tx.Hash().Hex())
	return execSuccessful(ctx, a.ContractBackend, tx)
}

func (a *Adjudicator) refute(ctx context.Context, req channel.AdjudicatorReq) error {
	ethParams := channelParamsToEthParams(req.Params)
	ethState := channelStateToEthState(req.Tx.State)
	a.mu.Lock()
	defer a.mu.Unlock()
	trans, err := a.newTransactor(ctx, big.NewInt(0), GasLimit)
	if err != nil {
		return errors.WithMessage(err, "creating transactor")
	}
	tx, err := a.contract.Refute(trans, ethParams, ethState, req.Tx.Sigs)
	if err != nil {
		return errors.Wrap(err, "calling concludeFinal")
	}
	log.Debugf("Sending transaction to the blockchain with txHash: %v", tx.Hash().Hex())
	return execSuccessful(ctx, a.ContractBackend, tx)
}
