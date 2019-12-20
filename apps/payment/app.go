// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

// Package payment implements the payment channel app.
package payment // import "perun.network/go-perun/apps/payment"

import (
	"io"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
)

// App is a payment app.
type App struct {
	Addr wallet.Address
}

// Def returns the address of this payment app.
func (a *App) Def() wallet.Address {
	return a.Addr
}

// DecodeData does not read anything from the reader and returns new NoData.
func (a *App) DecodeData(io.Reader) (channel.Data, error) {
	return new(NoData), nil
}

// ValidTransition checks that money flows only from the actor to the other
// participants.
func (a *App) ValidTransition(_ *channel.Params, from, to *channel.State, actor channel.Index) error {
	assertNoData(to)

	for i, bals := range from.OfParts {
		for j, bal := range bals {
			if int(actor) == i && bal.Cmp(to.OfParts[i][j]) == -1 {
				return errors.Errorf("payer[%d] steals asset %d", i, j)
			} else if int(actor) != i && bal.Cmp(to.OfParts[i][j]) == 1 {
				return errors.Errorf("payer[%d] reduces participant[%d]'s asset %d", actor, i, j)
			}
		}
	}
	return nil
}

// ValidInit panics if State.Data is not *NoData and returns nil otherwise. Any
// valid allocation forms a valid initial state.
func (a *App) ValidInit(_ *channel.Params, s *channel.State) error {
	assertNoData(s)
	return nil
}

func assertNoData(s *channel.State) {
	_, ok := s.Data.(*NoData)
	if !ok {
		log.Panicf("payment app must have no data (new(NoData)), has type %T", s.Data)
	}
}

// NoData represents empty app data.
type NoData struct{}

// Clone creates a new NoData
func (d *NoData) Clone() channel.Data {
	return new(NoData)
}

// Encode does nothing as NoData has no data.
func (d *NoData) Encode(io.Writer) error {
	return nil
}
