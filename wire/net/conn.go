// Copyright 2019 - See NOTICE file for copyright holders.
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
