// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

// A Bus is a central message bus over which all clients of a channel network
// communicate. It is used as the transport layer abstraction for the
// client.Client.
type Bus interface {
	Publisher

	// SubscribeClient should route all messages with clientAddr as recipient to
	// the provided Consumer. Every address may only be subscribed to once.
	SubscribeClient(c Consumer, clientAddr Address) error
}
