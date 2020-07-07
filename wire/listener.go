// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

// Listener is an interface that allows listening for peer incoming connections.
// The accepted connections still need to be authenticated.
type Listener interface {
	// Accept accepts an incoming connection, which still has to perform
	// authentication to exchange addresses.
	//
	// This function does not have to be reentrant, but concurrent calls to
	// Close() must abort ongoing Accept() calls. Accept() must only return
	// errors after Close() was called or an unrecoverable fatal error occurred
	// in the Listener and it is closed.
	Accept() (Conn, error)
	// Close closes the listener and aborts any ongoing Accept() call.
	Close() error
}
