// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package payment

import (
	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

// backend is set in init() to a new(Backend) and is used as a singleton.
var backend *Backend

// Backend is the payment app backend. The payment app's address has to be set
// once before using the app by calling SetAppDef().
type Backend struct {
	def wallet.Address
}

// AppFromDefinition returns a payment app if def matches the address set
// before and an error otherwise.
func (b *Backend) AppFromDefinition(def wallet.Address) (channel.App, error) {
	if b.def == nil {
		panic("def is nil")
	}

	if !b.def.Equals(def) {
		return nil, errors.Errorf("payment app has address %v, not %v", b.def, def)
	}

	return &App{def}, nil
}

// AppFromDefinition returns a payment app if def matches the address set
// before and an error otherwise.
func AppFromDefinition(def wallet.Address) (channel.App, error) {
	if backend.def == nil {
		panic("set the payment app's address once with SetAppDef before calling AppFromDefinition")
	}
	return backend.AppFromDefinition(def)
}

// SetAppDef sets the address of the payment app.
func (b *Backend) SetAppDef(def wallet.Address) {
	b.def = def
}

// SetAppDef sets the address of the payment app on the global app backend.
// The payment app's address must be set once at program start to the correct
// address with this function.
func SetAppDef(def wallet.Address) {
	backend.SetAppDef(def)
}

// AppDef gets the address of the payment app.
func (b *Backend) AppDef() wallet.Address {
	return b.def
}

// AppDef gets the address of the payment app of the global app backend.
func AppDef() wallet.Address {
	if backend.def == nil {
		panic("set the payment app's address once with SetAppDef before calling AppDef")
	}
	return backend.AppDef()
}
