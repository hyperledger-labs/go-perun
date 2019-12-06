// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package client

import (
	"context"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/sync/atomic"
)

type (
	// ChannelUpdate is a channel update proposal.
	ChannelUpdate struct {
		// State is the proposed new state.
		State *channel.State
		// ActorIdx is the actor causing the new state.  It does not need to
		// coincide with the sender of the request.
		ActorIdx uint16
	}

	UpdateHandler interface {
		Handle(ChannelUpdate, *UpdateResponder)
	}

	UpdateResponder struct {
		accept chan context.Context
		reject chan ctxUpdateRej
		err    chan error // return error
		called atomic.Bool
	}

	// The following type is only needed to bundle the ctx and channel update
	// rejection of UpdateResponder.Reject() into a single struct so that they can
	// be sent over a channel
	ctxUpdateRej struct {
		ctx    context.Context
		reason string
	}
)

func newUpdateResponder() *UpdateResponder {
	return &UpdateResponder{
		accept: make(chan context.Context),
		reject: make(chan ctxUpdateRej),
		err:    make(chan error, 1),
	}
}

// Accept lets the user signal that they want to accept the channel update.
func (r *UpdateResponder) Accept(ctx context.Context) error {
	if !r.called.TrySet() {
		log.Panic("multiple calls on channel update responder")
	}
	r.accept <- ctx
	return <-r.err
}

// Reject lets the user signal that they reject the channel update.
func (r *UpdateResponder) Reject(ctx context.Context, reason string) error {
	if !r.called.TrySet() {
		log.Panic("multiple calls on channel update responder")
	}
	r.reject <- ctxUpdateRej{ctx, reason}
	return <-r.err
}

}
