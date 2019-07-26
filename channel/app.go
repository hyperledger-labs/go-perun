// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"perun.network/go-perun/wallet"
)

type App interface {
	// Def is an identifier of the channel application. It is usually the
	// (counterfactual) on-chain address of the stateless contract that defines
	// what is a validTransition
	Def() wallet.Address

	// ValidTransition checks if the application specific rules of the given
	// transition from from to to are fulfilled.
	// The implementation should return a TransitionError describing the
	// invalidity of the transition, if it is not valid. It should return a normal
	// error (with attached stacktrace from pkg/errors) if there was any other
	// runtime error, not related to the invalidity of the transition itself.
	ValidTransition(parameters *Params, from, to *State) (bool, error)
}
