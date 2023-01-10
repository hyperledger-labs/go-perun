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

// Package payment implements the payment channel app.
package payment // import "perun.network/go-perun/apps/payment"

import (
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

// NewData returns a new instance of data specific to the payment app,
// intialized to its zero value.
//
// This should be used for unmarshalling the data from its binary
// representation.
func (a *App) NewData() channel.Data {
	return Data()
}

// ValidTransition checks that money flows only from the actor to the other
// participants.
func (a *App) ValidTransition(_ *channel.Params, from, to *channel.State, actor channel.Index) error {
	assertNoData(to)
	for i, asset := range from.Balances {
		for j, bal := range asset {
			if int(actor) == j && bal.Cmp(to.Balances[i][j]) == -1 {
				return errors.Errorf("payer[%d] steals asset %d, so %d < %d", j, i, bal, to.Balances[i][j])
			} else if int(actor) != j && bal.Cmp(to.Balances[i][j]) == 1 {
				return errors.Errorf("payer[%d] reduces participant[%d]'s asset %d", actor, j, i)
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
	if !channel.IsNoData(s.Data) {
		log.Panicf("payment app data must be NoData, is %T", s.Data)
	}
}

// Data returns the data that has to be used while creating a payment channel
// proposal. The Data is empty.
func Data() channel.Data {
	return channel.NoData()
}

// IsData returns wether an app data is valid payment app data.
func IsData(data channel.Data) bool {
	return channel.IsNoData(data)
}
