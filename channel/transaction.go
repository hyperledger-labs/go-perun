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
	"io"

	"github.com/pkg/errors"

	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire/perunio"
)

// Transaction is a channel state together with valid signatures from the
// channel participants.
type Transaction struct {
	*State
	Sigs []wallet.Sig
}

var _ perunio.Serializer = (*Transaction)(nil)

// Clone returns a deep copy of Transaction.
func (t Transaction) Clone() Transaction {
	return Transaction{
		State: t.State.Clone(),
		Sigs:  wallet.CloneSigs(t.Sigs),
	}
}

// Encode encodes a transaction into an `io.Writer` or returns an `error`.
func (t Transaction) Encode(w io.Writer) error {
	// Encode stateSet == 0
	if t.State == nil {
		return perunio.Encode(w, uint8(0))
	}

	// Encode stateSet and state
	if err := perunio.Encode(w, uint8(1), t.State); err != nil {
		return errors.WithMessage(err, "encoding stateSet bit and State")
	}
	return wallet.EncodeSparseSigs(w, t.Sigs)
}

// Decode decodes a transaction from an `io.Reader` or returns an `error`.
func (t *Transaction) Decode(r io.Reader) error {
	// Decode stateSet
	var stateSet uint8
	if err := perunio.Decode(r, &stateSet); err != nil {
		return errors.WithMessage(err, "decoding stateSet bit")
	}
	switch stateSet {
	case 0:
		t.State = nil
		return nil
	case 1:
	default:
		return errors.Errorf("unknown stateSet value: %v", stateSet)
	}

	// Decode State
	t.State = new(State)
	if err := perunio.Decode(r, t.State); err != nil {
		return errors.WithMessage(err, "decoding state")
	}

	t.Sigs = make([]wallet.Sig, t.State.NumParts())

	return wallet.DecodeSparseSigs(r, &t.Sigs)
}
