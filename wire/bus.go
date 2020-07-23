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
