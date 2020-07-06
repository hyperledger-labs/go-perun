// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

var _ Consumer = (*Relay)(nil)

// Relay is both a Consumer and a producer, and can be used to relay messages
// in a complex flow to Receivers.
type Relay struct {
	producer
}

// NewRelay initialises a new relay.
func NewRelay() *Relay {
	return &Relay{makeProducer()}
}

// Put puts a message into the relay.
func (r *Relay) Put(p *Peer, msg Msg) {
	r.produce(msg, p)
}
