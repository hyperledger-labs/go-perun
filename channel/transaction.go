// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package channel // import "perun.network/go-perun/channel"

import (
	"io"

	"github.com/pkg/errors"

	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

// Transaction is a channel state together with valid signatures from the
// channel participants.
type Transaction struct {
	*State
	Sigs []wallet.Sig
}

var _ perunio.Serializer = (*Transaction)(nil)

// Clone returns a deep copy of Transaction
func (t Transaction) Clone() Transaction {
	return Transaction{
		State: t.State.Clone(),
		Sigs:  wallet.CloneSigs(t.Sigs),
	}
}

// Encode encodes a transaction into an `io.Writer` or returns an `error`
func (t Transaction) Encode(w io.Writer) error {
	// Encode stateSet == 0
	if t.State == nil {
		return wire.Encode(w, uint8(0))
	}

	// Encode stateSet and state
	if err := wire.Encode(w, uint8(1), t.State); err != nil {
		return errors.WithMessage(err, "encoding stateSet bit and State")
	}
	return wallet.EncodeSparseSigs(w, t.Sigs)
}

// Decode decodes a transaction from an `io.Reader` or returns an `error`
func (t *Transaction) Decode(r io.Reader) error {
	// Decode stateSet
	var stateSet uint8
	if err := wire.Decode(r, &stateSet); err != nil {
		return errors.WithMessage(err, "decoding stateSet bit")
	}
	if (stateSet % 2) == 0 {
		t.State = nil
		return nil
	}

	// Decode State
	t.State = new(State)
	if err := wire.Decode(r, t.State); err != nil {
		return errors.WithMessage(err, "decoding state")
	}

	t.Sigs = make([]wallet.Sig, t.State.NumParts())

	return wallet.DecodeSparseSigs(r, &t.Sigs)
}
