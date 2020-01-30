// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import "context"

import "perun.network/go-perun/channel"

import "errors"

func (a *Adjudicator) register(ctx context.Context, request channel.AdjudicatorReq) error {
	return errors.New("not implemented")
}

func (a *Adjudicator) refute(ctx context.Context, request channel.AdjudicatorReq) error {
	return errors.New("not implemented")
}
