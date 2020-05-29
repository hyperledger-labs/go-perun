// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package keyvalue

import (
	"github.com/pkg/errors"

	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/pkg/sortedkv"
)

var _ persistence.PersistRestorer = (*PersistRestorer)(nil)

// PersistRestorer implements both the persister and the restorer interface
// using a sorted key-value store.
type PersistRestorer struct {
	cache *channelCache
	db    sortedkv.Database
}

// Close closes the PersistRestorer and releases all resources it holds.
func (pr *PersistRestorer) Close() error {
	if err := pr.db.Close(); err != nil {
		return err
	}
	pr.cache.clear()
	return nil
}

// NewPersistRestorer creates a new PersistRestorer for the supplied database.
func NewPersistRestorer(db sortedkv.Database) (*PersistRestorer, error) {
	r := &PersistRestorer{
		cache: newChannelCache(),
		db:    db,
	}

	return r, errors.WithMessage(r.readAllPeers(), "reading peers")
}
