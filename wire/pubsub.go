// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import "context"

type (
	// A Publisher allows to publish a message in a messaging network.
	Publisher interface {
		// Publish should return nil when the message was delivered (outgoing) or is
		// guaranteed to be eventually delivered (cached), depending on the goal of the
		// implementation.
		Publish(context.Context, *Envelope) error
	}

	// A subscriber allows to subscribe Consumers, which will receive messages
	// that match a predicate.
	Subscriber interface {
		// Subscribe adds a Consumer to the subscriptions.
		// If the Consumer is already subscribed, Subscribe panics.
		Subscribe(Consumer, Predicate) error
	}
)
