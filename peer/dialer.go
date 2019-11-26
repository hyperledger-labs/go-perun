// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"context"
)

// Dialer is an interface that allows creating a connection to a peer via its
// Perun address. The established connections are not authenticated yet.
type Dialer interface {
	// Dial creates a connection to a peer.
	// The passed context is used to abort the dialing process. The returned
	// connection might not belong to the requested address.
	//
	// Dial needs to be reentrant, and concurrent calls to Close() must abort
	// any ongoing Dial() calls.
	Dial(ctx context.Context, addr Address) (Conn, error)
	// Close aborts any ongoing calls to Dial().
	//
	// Close() needs to be reentrant, and repeated calls to Close() need to
	// return an error.
	Close() error
}
