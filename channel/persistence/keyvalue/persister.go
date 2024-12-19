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
	"bytes"
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/perunio"
	"polycry.pt/poly-go/sortedkv"
)

// ChannelCreated inserts a channel into the database.
func (pr *PersistRestorer) ChannelCreated(_ context.Context, s channel.Source, peers []map[wallet.BackendID]wire.Address, parent *map[wallet.BackendID]channel.ID) error {
	db := pr.channelDB(s.ID()).NewBatch()
	// Write the channel data in the "Channel" table.
	numParts := len(s.Params().Parts)
	keys := append([]string{"current", "index", "params", "phase", "staging:state"},
		sigKeys(numParts)...)
	if err := dbPutSource(db, s, keys...); err != nil {
		return err
	}

	if err := dbPut(db, "parent", optChannelIDEnc{parent}); err != nil {
		return errors.WithMessage(err, "putting parent ID")
	}

	// Write peers in the "Channel" table.
	if err := dbPut(db, prefix.Peers, (*wire.AddressMapArray)(&peers)); err != nil {
		return errors.WithMessage(err, "putting peers into channel table")
	}

	// Register the channel in the "Peer" table.
	peerdb := sortedkv.NewTable(pr.db, prefix.PeerDB).NewBatch()
	for _, peer := range peers {
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

// sigKey creates a key for given idx and number of channel
// participants.
func sigKey(idx, numParts int) string {
	width := int(math.Ceil(math.Log10(float64(numParts))))
	return fmt.Sprintf("%s%0*d", prefix.SigKey, width, idx)
}

// ChannelRemoved deletes a channel from the database.
func (pr *PersistRestorer) ChannelRemoved(ctx context.Context, id map[wallet.BackendID]channel.ID) error {
	db := pr.channelDB(id).NewBatch()
	peerdb := sortedkv.NewTable(pr.db, prefix.PeerDB).NewBatch()
	// All keys a channel has.
	params, err := pr.paramsForChan(id)
	if err != nil {
		return err
	}
	keys := append([]string{"current", "index", "params", "peers", "phase", "staging:state"},
		sigKeys(len(params.Parts))...)

	for _, key := range keys {
		if err := db.Delete(key); err != nil {
			return errors.WithMessage(err, "batch deletion of "+key)
		}
	}

	peers, err := pr.channelPeers(id)
	if err != nil {
		return errors.WithMessage(err, "retrieving peers for channel")
	}
	for _, peer := range peers {
		key, err := peerChannelKey(peer, id)
		if err != nil {
			return err
		}
		if err := peerdb.Delete(key); err != nil {
			return errors.WithMessage(err, "deleting peer channel")
		}
	}

	if err := db.Apply(); err != nil {
		return errors.WithMessage(err, "applying channel batch")
	}
	return errors.WithMessage(peerdb.Apply(), "applying peer batch")
}

// paramsForChan returns the channel parameters for a given channel id from
// the db.
func (pr *PersistRestorer) paramsForChan(id map[wallet.BackendID]channel.ID) (channel.Params, error) {
	params := channel.Params{}
	b, err := pr.channelDB(id).GetBytes("params")
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
		keys[i] = fmt.Sprintf(prefix.SigKey+"%0*d", width, i)
	}
	return keys
}

// Staged persists the staging transaction as well as the channel's phase.
func (pr *PersistRestorer) Staged(_ context.Context, s channel.Source) error {
	db := pr.channelDB(s.ID()).NewBatch()

	if err := dbPutSource(db, s, "staging:state", "phase"); err != nil {
		return err
	}

	return errors.WithMessage(db.Apply(), "applying batch")
}

// SigAdded persists the channel's staging transaction.
func (pr *PersistRestorer) SigAdded(_ context.Context, s channel.Source, idx channel.Index) error {
	db := pr.channelDB(s.ID()).NewBatch()

	numParts := len(s.Params().Parts)
	key := sigKey(int(idx), numParts)
	if err := dbPutSource(db, s, key); err != nil {
		return err
	}

	return errors.WithMessage(db.Apply(), "applying batch")
}

// Enabled persists the channel's staging and current transaction, and phase.
func (pr *PersistRestorer) Enabled(_ context.Context, s channel.Source) error {
	db := pr.channelDB(s.ID()).NewBatch()

	numParts := len(s.Params().Parts)
	keys := append([]string{"staging:state", "current", "phase"}, sigKeys(numParts)...)
	if err := dbPutSource(db, s, keys...); err != nil {
		return err
	}
	return errors.WithMessage(db.Apply(), "applying batch")
}

// PhaseChanged persists the channel's phase.
func (pr *PersistRestorer) PhaseChanged(_ context.Context, s channel.Source) error {
	return dbPut(pr.channelDB(s.ID()), "phase", s.Phase())
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

var sigRegex = regexp.MustCompile(`^` + prefix.SigKey + `\d+$`)

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
// form : "Colon:Separated:Key:IDX".
func decodeIdxFromDBKey(key string) (int, error) {
	vals := strings.Split(key, ":")
	return strconv.Atoi(vals[len(vals)-1])
}

func peerChannelKey(p map[wallet.BackendID]wire.Address, ch map[wallet.BackendID]channel.ID) (string, error) {
	var key bytes.Buffer
	if err := perunio.Encode(&key, wire.AddressDecMap(p)); err != nil {
		return "", errors.WithMessage(err, "encoding peer address")
	}
	key.WriteString(":channel:")
	if err := perunio.Encode(&key, channel.IDMap(ch)); err != nil {
		return "", errors.WithMessage(err, "encoding channel id")
	}
	return key.String(), nil
}

// channelDB creates a prefixed database for persisting a channel's data.
func (pr *PersistRestorer) channelDB(id map[wallet.BackendID]channel.ID) sortedkv.Database {
	return sortedkv.NewTable(pr.db, prefix.ChannelDB+channel.IDKey(id)+":")
}
