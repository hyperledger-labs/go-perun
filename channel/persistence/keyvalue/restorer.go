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
	"strings"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/perunio"
	"polycry.pt/poly-go/sortedkv"
)

var _ persistence.ChannelIterator = (*ChannelIterator)(nil)

// ChannelIterator implements the persistence.ChannelIterator interface.
type ChannelIterator struct {
	err error
	ch  *persistence.Channel
	its []sortedkv.Iterator

	restorer *PersistRestorer
}

// ActivePeers returns a list of all peers with which a channel is persisted.
func (pr *PersistRestorer) ActivePeers(ctx context.Context) ([]map[int]wire.Address, error) {
	it := sortedkv.NewTable(pr.db, prefix.PeerDB).NewIterator()

	peermap := make(map[wire.AddrKey]map[int]wire.Address)
	for it.Next() {
		var addr map[int]wire.Address
		err := perunio.Decode(bytes.NewBufferString(it.Key()), (*wire.AddressDecMap)(&addr))
		if err != nil {
			return nil, errors.WithMessagef(err, "decoding peer key (%x)", it.Key())
		}
		peermap[wire.Keys(addr)] = addr
	}

	peers := make([]map[int]wire.Address, 0, len(peermap))
	for _, peer := range peermap {
		peers = append(peers, peer)
	}
	return peers, errors.WithMessage(it.Close(), "closing iterator")
}

// channelPeers returns a slice of peer addresses for a given channel id from
// the db of PersistRestorer.
func (pr *PersistRestorer) channelPeers(id channel.ID) ([]map[int]wire.Address, error) {
	var ps wire.AddressMapArray
	peers, err := pr.channelDB(id).Get(prefix.Peers)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to get peerlist from db")
	}
	return ps, errors.WithMessage(perunio.Decode(bytes.NewBuffer([]byte(peers)), &ps),
		"decoding peerlist")
}

// RestoreAll should return an iterator over all persisted channels.
func (pr *PersistRestorer) RestoreAll() (persistence.ChannelIterator, error) {
	return &ChannelIterator{
		restorer: pr,
		its:      []sortedkv.Iterator{sortedkv.NewTable(pr.db, prefix.ChannelDB).NewIterator()},
	}, nil
}

// RestorePeer should return an iterator over all persisted channels which
// the given peer is a part of.
func (pr *PersistRestorer) RestorePeer(addr map[int]wire.Address) (persistence.ChannelIterator, error) {
	it := &ChannelIterator{restorer: pr}
	chandb := sortedkv.NewTable(pr.db, prefix.ChannelDB)

	key, err := peerChannelsKey(addr)
	if err != nil {
		return nil, errors.WithMessage(err, "restoring peer")
	}
	itPeer := sortedkv.NewTable(pr.db, prefix.PeerDB+key).NewIterator()
	defer itPeer.Close()

	var id channel.ID
	for itPeer.Next() {
		if err := perunio.Decode(bytes.NewBufferString(itPeer.Key()), &id); err != nil {
			return nil, errors.WithMessage(err, "decode channel id")
		}
		it.its = append(it.its, chandb.NewIteratorWithPrefix(string(id[:])))
	}

	return it, nil
}

// peerChannelsKey creates a db-key-string for a given wire.Address.
func peerChannelsKey(addr map[int]wire.Address) (string, error) {
	var key strings.Builder
	if err := perunio.Encode(&key, wire.AddressDecMap(addr)); err != nil {
		return "", errors.WithMessage(err, "encoding peer address")
	}
	key.WriteString(":channel:")
	return key.String(), nil
}

// RestoreChannel restores a single channel.
func (pr *PersistRestorer) RestoreChannel(ctx context.Context, id channel.ID) (*persistence.Channel, error) {
	chandb := sortedkv.NewTable(pr.db, prefix.ChannelDB)
	it := &ChannelIterator{
		restorer: pr,
		its:      []sortedkv.Iterator{chandb.NewIteratorWithPrefix(string(id[:]))},
	}

	if it.Next(ctx) {
		return it.Channel(), it.Close()
	}
	if err := it.Close(); err != nil {
		return nil, errors.WithMessagef(err, "error restoring channel %x", id)
	}
	return nil, errors.Errorf("could not find channel %x", id)
}

type decOpts uint8

const (
	noOpts   decOpts = 0
	allowEnd decOpts = 1 << iota
	allowEmpty
)

// isSetIn masks the given opts with the current opt and returns
// wether or not the corresponding bit of opt is set within opts.
func (opt decOpts) isSetIn(opts decOpts) bool {
	return opt&opts == opt
}

// Next advances the iterator and returns whether there is another channel.
func (i *ChannelIterator) Next(context.Context) bool {
	if len(i.its) == 0 {
		return false
	}

	i.ch = persistence.NewChannel()
	if !i.decodeNext("current", &i.ch.CurrentTXV, allowEnd) ||
		!i.decodeNext("index", &i.ch.IdxV, noOpts) ||
		!i.decodeNext("params", i.ch.ParamsV, noOpts) ||
		!i.decodeNext("parent", optChannelIDDec{&i.ch.Parent}, noOpts) ||
		!i.decodeNext("peers", (*wire.AddressMapArray)(&i.ch.PeersV), noOpts) ||
		!i.decodeNext("phase", &i.ch.PhaseV, noOpts) {
		return false
	}
	i.ch.StagingTXV.Sigs = make([]wallet.Sig, len(i.ch.ParamsV.Parts))
	for idx, key := range sigKeys(len(i.ch.ParamsV.Parts)) {
		i.decodeNext(key, wallet.SigDec{Sig: &i.ch.StagingTXV.Sigs[idx]}, allowEmpty)
	}

	return i.decodeNext("staging:state", &PersistedState{&i.ch.StagingTXV.State}, allowEmpty)
}

// recoverFromEmptyIterator is called when there is no iterator or when the
// current iterator just ended. allowEnd signifies whether this situation is
// allowed. It returns whether it could recover a new iterator.
func (i *ChannelIterator) recoverFromEmptyIterator(key string, allowedToEnd decOpts) bool {
	err := i.its[0].Close()
	i.its = i.its[1:]
	if err != nil {
		i.err = err
		return false
	}

	if !allowEnd.isSetIn(allowedToEnd) {
		i.err = errors.New("iterator ended; expected key " + key)
		return false
	}

	return len(i.its) != 0
}

// decodeNext reduces code duplication for decoding a value from an iterator. If
// an iterator ends in the middle of decoding a channel, then the channel
// iterator's error is set. Returns whether a value was decoded without error.
func (i *ChannelIterator) decodeNext(key string, v interface{}, opts decOpts) bool {
	for !i.its[0].Next() {
		if !i.recoverFromEmptyIterator(key, opts) {
			return false
		}
	}

	buf := bytes.NewBuffer(i.its[0].ValueBytes())
	if buf.Len() == 0 {
		if allowEmpty.isSetIn(opts) {
			return true
		}
		i.err = errors.Errorf("unexpected empty value")
		return false
	}

	i.err = errors.WithMessage(perunio.Decode(buf, v), "decoding "+key)
	if i.err != nil {
		return false
	}
	if buf.Len() != 0 {
		i.err = errors.Errorf("decoding %s incomplete (%d bytes left)", key, buf.Len())
	}

	return i.err == nil
}

// Channel returns the iterator's current channel.
func (i *ChannelIterator) Channel() *persistence.Channel {
	return i.ch
}

// Close closes the iterator and releases its resources. It returns the last
// error that occurred when advancing the iterator.
func (i *ChannelIterator) Close() error {
	for it := range i.its {
		if err := i.its[it].Close(); err != nil && i.err == nil {
			i.err = err
		}
	}
	i.its = nil

	return i.err
}
