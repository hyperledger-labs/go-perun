// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	wire "perun.network/go-perun/wire/msg"
)

var _ Consumer = (*Relay)(nil)

// Relay is both a Consumer and a producer, and can be used to relay messages
// in a complex flow to Receivers.
type Relay struct {
	producer
}

// MakeRelay initialises a new relay.
func NewRelay() *Relay {
	return &Relay{makeProducer()}
}

func (r *Relay) Put(p *Peer, msg wire.Msg) {
	r.produce(msg, p)
}
