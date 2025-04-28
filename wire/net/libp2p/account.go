// Copyright 2025 - See NOTICE file for copyright holders.
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

package libp2p

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
)

const (
	relayID = "QmVCPfUMr98PaaM8qbAQBgJ9jqc7XHpGp7AsyragdFDmgm"

	queryProtocol    = "/address-book/query/1.0.0"    // Protocol for querying the relay-server for a peerID.
	registerProtocol = "/address-book/register/1.0.0" // Protocol for registering an on-chain address with the relay-server.
	removeProtocol   = "/address-book/remove/1.0.0"   // Protocol for deregistering an on-chain address with the relay-server.
)

// Account represents a libp2p wire account containting a libp2p host.
type Account struct {
	host.Host
	relayAddr  string
	privateKey crypto.PrivKey
}
