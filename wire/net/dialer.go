// Copyright 2020 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package net // import "perun.network/go-perun/wire/net"

import (
	"context"

	"perun.network/go-perun/wire"
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
	Dial(ctx context.Context, addr wire.Address) (Conn, error)
	// Close aborts any ongoing calls to Dial().
	//
	// Close() needs to be reentrant, and repeated calls to Close() need to
	// return an error.
	Close() error
}
