// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package net

import "perun.network/go-perun/wire"

// Conn is a connection to a peer, and can send wire messages.
// The Send and Recv methods do not have to be reentrant, but calls to Close
// that happen in other threads must interrupt ongoing Send and Recv calls.
// This is the default behavior for sockets.
type Conn interface {
	// Recv receives an envelope from the peer.
	// If an error occurs, the connection must close itself.
	Recv() (*wire.Envelope, error)
	// Send sends an envelope to the peer.
	// If an error occurs, the connection must close itself.
	Send(*wire.Envelope) error
	// Close closes the connection and aborts any ongoing Send() and Recv()
	// calls.
	//
	// Repeated calls to Close() result in an error.
	Close() error
}
