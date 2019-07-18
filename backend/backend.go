// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package backend provides functionalities to set the global backend for the
// go-perun framework. It must only be set through backend.Set and not on the
// other package's SetBackend() functions, which are called by this function.
package backend // import "perun.network/go-perun/backend"

import (
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
)

// isSet stores whether the global backend has already been set. Relevant for
// Set(Collection).
var isSet = false

// Collection encapsulates multiple interfaces that can be used by the core
// functionality. It should contain all helper interfaces of the subpackages.
type Collection struct {
	Channel channel.Backend
	Wallet  wallet.Backend
}

// Set sets the global backend. It must only be called once and panics
// otherwise.
func Set(c Collection) {
	if isSet {
		log.Panic("Backend can only be set once.")
	}
	channel.SetBackend(c.Channel)
	wallet.SetBackend(c.Wallet)
	isSet = true
}
