// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package keyvalue

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/pkg/sortedkv"
	"perun.network/go-perun/wire"
)

// ChannelCreated inserts a channel into the database.
func (p *PersistRestorer) ChannelCreated(_ context.Context, s channel.Source, peers []wire.Address) error {
	db := p.channelDB(s.ID()).NewBatch()
	// Write the channel data in the "Channel" table.
	numParts := len(s.Params().Parts)
	keys := append([]string{"current", "index", "params", "phase", "staging:state"}, sigKeys(numParts)...)
	if err := dbPutSource(db, s, keys...); err != nil {
		return err
	}

	// Register the channel in the "Peer" table.
	peerdb := sortedkv.NewTable(p.db, prefix.PeerDB).NewBatch()
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

const dbSigKey = "staging:sig:"

// sigKey creates a key for given idx and number of channel
// participants.
func sigKey(idx, numParts int) string {
	width := int(math.Ceil(math.Log10(float64(numParts))))
	return fmt.Sprintf("%s%0*d", dbSigKey, width, idx)
}

// ChannelRemoved deletes a channel from the database.
func (p *PersistRestorer) ChannelRemoved(_ context.Context, id channel.ID) error {
	db := p.channelDB(id).NewBatch()
	peerdb := sortedkv.NewTable(p.db, prefix.PeerDB).NewBatch()
	// All keys a channel has.
	params, err := getParamsForChan(p.db, id)
	if err != nil {
		return err
	}
	keys := append([]string{"current", "index", "params", "phase", "staging:state"}, sigKeys(len(params.Parts))...)

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

// getParamsForChan returns the channel parameters for a given channel id from
// the db.
func getParamsForChan(db sortedkv.Reader, id channel.ID) (channel.Params, error) {
	params := channel.Params{}
	b, err := db.GetBytes("Chan:" + string(id[:]) + ":params")
	if err != nil {
		return params, errors.WithMessage(err, "unable to retrieve params from db")
	}
	return params, errors.WithMessage(perunio.Decode(bytes.NewBuffer(b), &params),
		"unable to decode channel parameters")
}

// sigKeys generates all db keys for signatures and returns them as a
// slice of strings.
func sigKeys(numParts int) []string {
	keys := make([]string, numParts)
	width := int(math.Ceil(math.Log10(float64(numParts))))
	for i := range keys {
		keys[i] = fmt.Sprintf(dbSigKey+"%0*d", width, i)
	}
	return keys
}

// Staged persists the staging transaction as well as the channel's phase.
func (p *PersistRestorer) Staged(_ context.Context, s channel.Source) error {
	db := p.channelDB(s.ID()).NewBatch()

	if err := dbPutSource(db, s, "staging:state", "phase"); err != nil {
		return err
	}

	return errors.WithMessage(db.Apply(), "applying batch")
}

// SigAdded persists the channel's staging transaction.
func (p *PersistRestorer) SigAdded(_ context.Context, s channel.Source, idx channel.Index) error {
	db := p.channelDB(s.ID()).NewBatch()

	numParts := len(s.Params().Parts)
	key := sigKey(int(idx), numParts)
	dbPutSource(db, s, key)

	return errors.WithMessage(db.Apply(), "applying batch")
}

// Enabled persists the channel's staging and current transaction, and phase.
func (p *PersistRestorer) Enabled(_ context.Context, s channel.Source) error {
	db := p.channelDB(s.ID()).NewBatch()

	numParts := len(s.Params().Parts)
	keys := append([]string{"staging:state", "current", "phase"}, sigKeys(numParts)...)
	if err := dbPutSource(db, s, keys...); err != nil {
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
	case "staging:state":
		stagingState := s.StagingTX().State
		return dbPut(db, key, PersistedState{&stagingState})
	}
	if idx, ok := sigKeyIndex(key); ok {
		tx := s.StagingTX()
		if len(tx.Sigs) == 0 || len(tx.Sigs) <= idx {
			return dbPut(db, key, []byte(""))
		}
		return dbPut(db, key, tx.Sigs[idx])
	}
	panic("unknown key: " + key)
}

// dbPut reduces code duplication for encoding and writing data to a database.
func dbPut(db sortedkv.Writer, key string, v interface{}) error {
	var buf bytes.Buffer
	if err := perunio.Encode(&buf, v); err != nil {
		return errors.WithMessage(err, "encoding "+key)
	}
	if err := db.PutBytes(key, buf.Bytes()); err != nil {
		return errors.WithMessage(err, "putting "+key)
	}

	return nil
}

var sigRegex = regexp.MustCompile(`^` + dbSigKey + `\d+$`)

func sigKeyIndex(key string) (int, bool) {
	if !sigRegex.MatchString(key) {
		return -1, false
	}
	idx, err := decodeIdxFromDBKey(key)
	if err != nil {
		return -1, false
	}
	return idx, true
}

// decodeIdxFromDBKey decodes the encoded idx within a db key of the following
// form : "Colon:Separated:Key:IDX"
func decodeIdxFromDBKey(key string) (int, error) {
	vals := strings.Split(key, ":")
	return strconv.Atoi(vals[len(vals)-1])
}

func peerChannelKey(p wire.Address, ch channel.ID) (string, error) {
	var key bytes.Buffer
	if err := p.Encode(&key); err != nil {
		return "", errors.WithMessage(err, "encoding peer address")
	}
	key.WriteString(":channel:")
	if err := perunio.Encode(&key, ch); err != nil {
		return "", errors.WithMessage(err, "encoding channel id")
	}
	return key.String(), nil
}

// channelDB creates a prefixed database for persisting a channel's data.
func (p *PersistRestorer) channelDB(id channel.ID) sortedkv.Database {
	return sortedkv.NewTable(p.db, prefix.ChannelDB+string(id[:])+":")
}
