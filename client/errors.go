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

package client

import (
	"fmt"

	"github.com/pkg/errors"
)

type (
	// TxTimedoutError indicates that we have timed out waiting for a
	// transaction to be mined.
	//
	// It can happen that this transaction can be eventually mined. So the user
	// should of the framework should be watching for the transaction to be
	// mined and decide on what to do.
	TxTimedoutError struct {
		TxType string // Type of the transaction.
		TxID   string // Transaction ID to track it on the blockchain.
	}

	// ChainNotReachableError indicates problems in connecting to the blockchain
	// network when trying to do on-chain transactions or reading from the blockchain.
	ChainNotReachableError struct {
	}
)

// Error implements the error interface.
func (e TxTimedoutError) Error() string {
	return fmt.Sprintf("timed out waiting for tx to be mined. txID: %s, TxType: %s", e.TxID, e.TxType)
}

// Error implements the error interface.
func (e ChainNotReachableError) Error() string {
	return "blockchain network not reachable"
}

// NewTxTimedoutError constructs a TxTimedoutError and wraps it with the actual
// error message.
//
// txID is the ID required for tracking the transaction on the blockchain and
// txType is the type of on-chain transaction. Valid types should be defined by
// each of the blockchain backend.
func NewTxTimedoutError(txType, txID, actualErrMsg string) error {
	return errors.Wrap(TxTimedoutError{
		TxType: txType,
		TxID:   txID,
	}, actualErrMsg)
}

// NewChainNotReachableError constructs a ChainNotReachableError and wraps it
// with the actual error message.
func NewChainNotReachableError(actualErr error) error {
	return errors.Wrap(ChainNotReachableError{}, actualErr.Error())
}
