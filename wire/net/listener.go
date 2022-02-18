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

package net

import "perun.network/go-perun/wire"

// Listener is an interface that allows listening for peer incoming connections.
// The accepted connections still need to be authenticated.
type Listener interface {
	// Accept accepts an incoming connection, which still has to perform
	// authentication to exchange addresses.
	//
	// `ser` specifies the message serialization format.
	//
	// This function does not have to be reentrant, but concurrent calls to
	// Close() must abort ongoing Accept() calls. Accept() must only return
	// errors after Close() was called or an unrecoverable fatal error occurred
	// in the Listener and it is closed.
	Accept(ser wire.EnvelopeSerializer) (Conn, error)
	// Close closes the listener and aborts any ongoing Accept() call.
	Close() error
}
