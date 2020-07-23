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

package keyvalue

import (
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/pkg/sortedkv"
)

var _ persistence.PersistRestorer = (*PersistRestorer)(nil)

// PersistRestorer implements both the persister and the restorer interface
// using a sorted key-value store.
type PersistRestorer struct {
	db sortedkv.Database
}

// Close closes the PersistRestorer and releases all resources it holds.
func (pr *PersistRestorer) Close() error {
	if err := pr.db.Close(); err != nil {
		return err
	}
	return nil
}

// NewPersistRestorer creates a new PersistRestorer for the supplied database.
func NewPersistRestorer(db sortedkv.Database) *PersistRestorer {
	return &PersistRestorer{
		db: db,
	}
}

var prefix = struct{ ChannelDB, PeerDB, SigKey, Peers string }{
	ChannelDB: "Chan:",
	PeerDB:    "Peer:",
	SigKey:    "staging:sig:",
	Peers:     "peers",
}
