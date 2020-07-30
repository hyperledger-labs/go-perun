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
	"bytes"
	"context"
	stderrors "errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"perun.network/go-perun/log"
)

// How many blocks we query into the past for events.
const startBlockOffset = 100

// GasLimit is the max amount of gas we want to send per transaction.
const GasLimit = 500000

// ContractInterface provides all functions needed by an ethereum backend.
// Both test.SimulatedBackend and ethclient.Client implement this interface.
type ContractInterface interface {
	bind.ContractBackend
	ethereum.ChainReader
	ethereum.TransactionReader
}

// ContractBackend adds a keystore and an on-chain account to the ContractInterface.
// This is needed to send on-chain transaction to interact with the smart contracts.
type ContractBackend struct {
	ContractInterface
	ks      *keystore.KeyStore
	account *accounts.Account
}

// NewContractBackend creates a new ContractBackend with the given parameters.
func NewContractBackend(cf ContractInterface, ks *keystore.KeyStore, acc *accounts.Account) ContractBackend {
	return ContractBackend{
		ContractInterface: cf,
		ks:                ks,
		account:           acc,
	}
}

// NewWatchOpts returns bind.WatchOpts with the field Start set to the current
// block number and the ctx field set to the passed context.
func (c *ContractBackend) NewWatchOpts(ctx context.Context) (*bind.WatchOpts, error) {
	blockNum, err := c.pastOffsetBlockNum(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "new watch opts")
	}

	return &bind.WatchOpts{
		Start:   &blockNum,
		Context: ctx,
	}, nil
}

// NewFilterOpts returns bind.FilterOpts with the field Start set to the block
// number 100 blocks ago (or 1) and the field End set to nil and the ctx field
// set to the passed context.
func (c *ContractBackend) NewFilterOpts(ctx context.Context) (*bind.FilterOpts, error) {
	blockNum, err := c.pastOffsetBlockNum(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "new filter opts")
	}
	return &bind.FilterOpts{
		Start:   blockNum,
		End:     nil,
		Context: ctx,
	}, nil
}

func (c *ContractBackend) pastOffsetBlockNum(ctx context.Context) (uint64, error) {
	h, err := c.HeaderByNumber(ctx, nil)
	if err != nil {
		return uint64(0), errors.Wrap(err, "retrieving latest block")
	}

	// max(1, latestBlock - offset)
	if h.Number.Uint64() <= startBlockOffset {
		return 1, nil
	}
	return h.Number.Uint64() - startBlockOffset, nil
}

// NewTransactor returns bind.TransactOpts with the current nonce, suggested gas
// price and account of the ContractBackend. The gasLimit and value in wei are
// taken from the parameters.
func (c *ContractBackend) NewTransactor(ctx context.Context, valueWei *big.Int, gasLimit uint64) (*bind.TransactOpts, error) {
	nonce, err := c.PendingNonceAt(ctx, c.account.Address)
	if err != nil {
		return nil, errors.Wrap(err, "querying pending nonce")
	}

	gasPrice, err := c.SuggestGasPrice(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "querying suggested gas price")
	}

	auth, err := bind.NewKeyStoreTransactor(c.ks, *c.account)
	if err != nil {
		return nil, errors.Wrap(err, "creating transactor")
	}

	auth.Nonce = new(big.Int).SetUint64(nonce)
	auth.Value = valueWei    // in wei
	auth.GasLimit = gasLimit // in units
	auth.GasPrice = gasPrice

	return auth, nil
}

func (c *ContractBackend) confirmTransaction(ctx context.Context, tx *types.Transaction) error {
	receipt, err := bind.WaitMined(ctx, c, tx)
	if err != nil {
		return errors.Wrap(err, "sending transaction")
	}
	if receipt.Status == types.ReceiptStatusFailed {
		reason, err := errorReason(ctx, c, tx, receipt.BlockNumber)
		if err != nil {
			log.Warn("TX failed; error determining reason: ", err)
		} else {
			log.Warn("TX failed with reason: ", reason)
		}
		return errors.WithStack(ErrorTxFailed)
	}
	return nil
}

// ErrorTxFailed signals a failed, i.e., reverted, transaction.
var ErrorTxFailed = stderrors.New("transaction failed")

// IsTxFailedError returns whether the cause of the error was a failed transaction.
func IsTxFailedError(err error) bool {
	return errors.Cause(err) == ErrorTxFailed
}

func errorReason(ctx context.Context, b *ContractBackend, tx *types.Transaction, blockNum *big.Int) (string, error) {
	msg := ethereum.CallMsg{
		From:     b.account.Address,
		To:       tx.To(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
		Data:     tx.Data(),
	}
	res, err := b.CallContract(ctx, msg, blockNum)
	if err != nil {
		return "", errors.Wrap(err, "CallContract")
	}
	return unpackError(res)
}

// Keccak256("Error(string)")[:4].
var errorSig = []byte{0x08, 0xc3, 0x79, 0xa0}

func unpackError(result []byte) (string, error) {
	if len(result) < 4 || !bytes.Equal(result[:4], errorSig) {
		return "<tx result not Error(string)>", errors.New("TX result not of type Error(string)")
	}
	vs, err := abi.Arguments{{Type: abiString}}.UnpackValues(result[4:])
	if err != nil {
		return "<invalid tx result>", errors.Wrap(err, "unpacking revert reason")
	}
	return vs[0].(string), nil
}

// ContractBytecodeError signals invalid bytecode at given address, such as incorrect or no code.
// nolint:stylecheck
var ContractBytecodeError = stderrors.New("invalid bytecode at address")

// IsContractBytecodeError returns whether the cause of the error was a invalid bytecode.
func IsContractBytecodeError(err error) bool {
	return errors.Cause(err) == ContractBytecodeError
}

// FetchCodeAtAddr reads the bytecode at given address.
// Returns a ContractBytecodeError when there is no bytecode at given address.
// This error can be checked with IsContractBytecodeError() function.
func FetchCodeAtAddr(ctx context.Context, backend ContractBackend, contractAddr common.Address) ([]byte, error) {
	code, err := backend.CodeAt(ctx, contractAddr, nil)
	if err != nil {
		return nil, err
	}
	if len(code) == 0 {
		return nil, errors.WithMessage(ContractBytecodeError, "no code")
	}
	return code, nil
}
