// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client_test

import (
	"context"
	"testing"
	"time"

	perunchannel "perun.network/go-perun/channel"
)

// DummyAdjudicator is a temporary dummy implementation of the Adjudicator interface until the Ethereum Adjudicator is merged
type DummyAdjudicator struct {
	t *testing.T
}

func (d *DummyAdjudicator) Register(ctx context.Context, req perunchannel.AdjudicatorReq) (*perunchannel.RegisteredEvent, error) {
	return &perunchannel.RegisteredEvent{
		ID:      req.Params.ID(),
		Idx:     req.Idx,
		Version: req.Tx.Version,
		Timeout: time.Now(),
	}, nil
}

func (d *DummyAdjudicator) Withdraw(context.Context, perunchannel.AdjudicatorReq) error {
	return nil
}

func (d *DummyAdjudicator) SubscribeRegistered(context.Context, *perunchannel.Params) (perunchannel.RegisteredSubscription, error) {
	return nil, nil
}
