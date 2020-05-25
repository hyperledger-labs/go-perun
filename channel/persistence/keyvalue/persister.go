// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package keyvalue

import (
	"bytes"
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/pkg/sortedkv"
	"perun.network/go-perun/wire"
)

// ChannelCreated inserts a channel into the database.
func (p *PersistRestorer) ChannelCreated(_ context.Context, s channel.Source, peers []peer.Address) error {
	db := p.channelDB(s.ID()).NewBatch()
	// Write the channel data in the "Channel" table.
	if err := dbPutSource(db, s,
		"current",
		"index",
		"params",
		"phase",
		"staging"); err != nil {
		return err
	}
	if err := dbPut(db, "peers", peer.Addresses(peers)); err != nil {
		return err
	}

	// Register the channel in the "Peer" table.
	peerdb := sortedkv.NewTable(p.db, "Peer:").NewBatch()
	for _, peer := range peers {
		p.cache.addPeerChannel(peer, s.ID())
		key, err := peerChannelKey(peer, s.ID())
		if err != nil {
			return err
		}
		if err := peerdb.Put(key, ""); err != nil {
			return errors.WithMessage(err, "putting peer channel")
		}
	}

	if err := db.Apply(); err != nil {
		return errors.WithMessage(err, "applying channel batch")
	}
	return errors.WithMessage(peerdb.Apply(), "applying peer batch")
}

// ChannelRemoved deletes a channel from the database.
func (p *PersistRestorer) ChannelRemoved(_ context.Context, id channel.ID) error {
	db := p.channelDB(id).NewBatch()
	peerdb := sortedkv.NewTable(p.db, "Peer:").NewBatch()
	// All keys a channel has.
	var keys = []string{"current", "index", "params", "peers", "phase", "staging"}

	for _, key := range keys {
		if err := db.Delete(key); err != nil {
			return errors.WithMessage(err, "batch deletion of "+key)
		}
	}

	for _, peer := range p.cache.peers[id] {
		key, err := peerChannelKey(peer, id)
		if err != nil {
			return err
		}
		if err := peerdb.Delete(key); err != nil {
			return errors.WithMessage(err, "deleting peer channel")
		}
	}

	p.cache.deleteChannel(id)

	if err := db.Apply(); err != nil {
		return errors.WithMessage(err, "applying channel batch")
	}
	return errors.WithMessage(peerdb.Apply(), "applying peer batch")
}

// Staged persists the staging transaction as well as the channel's phase.
func (p *PersistRestorer) Staged(_ context.Context, s channel.Source) error {
	db := p.channelDB(s.ID()).NewBatch()

	if err := dbPutSource(db, s, "staging", "phase"); err != nil {
		return err
	}

	return errors.WithMessage(db.Apply(), "applying batch")
}

// SigAdded persists the channel's staging transaction.
func (p *PersistRestorer) SigAdded(_ context.Context, s channel.Source, _ channel.Index) error {
	db := p.channelDB(s.ID()).NewBatch()

	if err := dbPut(db, "staging", s.StagingTX()); err != nil {
		return err
	}

	return errors.WithMessage(db.Apply(), "applying batch")
}

// Enabled persists the channel's staging and current transaction, and phase.
func (p *PersistRestorer) Enabled(_ context.Context, s channel.Source) error {
	db := p.channelDB(s.ID()).NewBatch()

	if err := dbPut(db, "staging", s.StagingTX()); err != nil {
		return err
	}

	if err := dbPut(db, "current", s.CurrentTX()); err != nil {
		return err
	}

	if err := dbPut(db, "phase", s.Phase()); err != nil {
		return err
	}

	return errors.WithMessage(db.Apply(), "applying batch")
}

// PhaseChanged persists the channel's phase.
func (p *PersistRestorer) PhaseChanged(_ context.Context, s channel.Source) error {
	return dbPut(p.channelDB(s.ID()), "phase", s.Phase())
}

func dbPutSource(db sortedkv.Writer, s channel.Source, keys ...string) error {
	for _, key := range keys {
		if err := dbPutSourceField(db, s, key); err != nil {
			return err
		}
	}
	return nil
}

func dbPutSourceField(db sortedkv.Writer, s channel.Source, key string) error {
	switch key {
	case "current":
		return dbPut(db, key, s.CurrentTX())
	case "index":
		return dbPut(db, key, s.Idx())
	case "params":
		return dbPut(db, key, s.Params())
	case "phase":
		return dbPut(db, key, s.Phase())
	case "staging":
		return dbPut(db, key, s.StagingTX())
	default:
		panic("unknown key: " + key)
	}
}

// dbPut reduces code duplication for encoding and writing data to a database.
func dbPut(db sortedkv.Writer, key string, v interface{}) error {
	var buf bytes.Buffer
	if err := wire.Encode(&buf, v); err != nil {
		return errors.WithMessage(err, "encoding "+key)
	}
	if err := db.PutBytes(key, buf.Bytes()); err != nil {
		return errors.WithMessage(err, "putting "+key)
	}

	return nil
}

func peerChannelKey(p peer.Address, ch channel.ID) (string, error) {
	var key bytes.Buffer
	if err := p.Encode(&key); err != nil {
		return "", errors.WithMessage(err, "encoding peer address")
	}
	key.WriteString(":channel:")
	if err := wire.Encode(&key, ch); err != nil {
		return "", errors.WithMessage(err, "encoding channel id")
	}
	return key.String(), nil
}

// channelDB creates a prefixed database for persisting a channel's data.
func (p *PersistRestorer) channelDB(id channel.ID) sortedkv.Database {
	return sortedkv.NewTable(p.db, "Chan:"+string(id[:])+":")
}
