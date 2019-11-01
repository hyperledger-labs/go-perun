// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

// Listener is an interface that allows listening for peer incoming connections.
// The accepted connections still need to be authenticated.
type Listener interface {
	// Accept accepts an incoming connection, which still has to perform
	// authentication to exchange addresses.
	//
	// This function does not have to be reentrant, but concurrent calls to
	// Close() must abort ongoing Accept() calls. Accept() must never return
	// errors except after Close() was called.
	Accept() (Conn, error)
	// Close closes the listener and aborts any ongoing Accept() call.
	Close() error
}
