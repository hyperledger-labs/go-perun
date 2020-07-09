// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package keyvalue

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/pkg/sortedkv"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
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
func (r *PersistRestorer) ActivePeers(context.Context) ([]wire.Address, error) {
	it := sortedkv.NewTable(r.db, prefix.PeerDB).NewIterator()

	peermap := make(map[wallet.AddrKey]wire.Address)
	for it.Next() {
		addr, err := wire.DecodeAddress(bytes.NewBufferString(it.Key()))
		if err != nil {
			return nil, errors.WithMessagef(err, "decoding peer key (%x)", it.Key())
		}
		peermap[wallet.Key(addr)] = addr
	}

	peers := make([]wire.Address, 0, len(peermap))
	for _, peer := range peermap {
		peers = append(peers, peer)
	}
	return peers, errors.WithMessage(it.Close(), "closing iterator")
}

// RestoreAll should return an iterator over all persisted channels.
func (r *PersistRestorer) RestoreAll() (persistence.ChannelIterator, error) {
	return &ChannelIterator{
		restorer: r,
		its:      []sortedkv.Iterator{sortedkv.NewTable(r.db, prefix.ChannelDB).NewIterator()},
	}, nil
}

// RestorePeer should return an iterator over all persisted channels which
// the given peer is a part of.
func (r *PersistRestorer) RestorePeer(addr wire.Address) (persistence.ChannelIterator, error) {
	it := &ChannelIterator{restorer: r}
	chandb := sortedkv.NewTable(r.db, prefix.ChannelDB)

	key, err := peerChannelsKey(addr)
	if err != nil {
		return nil, errors.WithMessage(err, "restoring peer")
	}
	itPeer := sortedkv.NewTable(r.db, prefix.PeerDB+key).NewIterator()
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
func peerChannelsKey(addr wire.Address) (string, error) {
	var key strings.Builder
	if err := addr.Encode(&key); err != nil {
		return "", errors.WithMessage(err, "encoding peer address")
	}
	key.WriteString(":channel:")
	return key.String(), nil
}

// RestoreChannel restores a single channel.
func (r *PersistRestorer) RestoreChannel(ctx context.Context, id channel.ID) (*persistence.Channel, error) {
	chandb := sortedkv.NewTable(r.db, prefix.ChannelDB)
	it := &ChannelIterator{
		restorer: r,
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

// decodePeerChanID decodes the channel.ID and peer.Address from a key.
func decodePeerChanID(key string) (wire.Address, channel.ID, error) {
	buf := bytes.NewBufferString(key)
	addr, err := wire.DecodeAddress(buf)
	if err != nil {
		return addr, channel.ID{}, errors.WithMessage(err, "decode peer address")
	}

	if err = eatExpect(buf, ":channel:"); err != nil {
		return nil, channel.ID{}, errors.WithMessagef(err, "key: %x", key)
	}

	var id channel.ID
	return addr, id, errors.WithMessage(perunio.Decode(buf, &id), "decode channel id")
}

// eatExpect consumes bytes from a Reader and asserts that they are equal to
// the expected string.
func eatExpect(r io.Reader, tok string) error {
	buf := make([]byte, len(tok))
	if _, err := io.ReadFull(r, buf); err != nil {
		return errors.WithMessage(err, "reading")
	}
	if string(buf) != tok {
		return errors.Errorf("expected %s, got %s.", tok, string(buf))
	}
	return nil
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

	i.ch = &persistence.Channel{ParamsV: new(channel.Params)}

	if !i.decodeNext("current", &i.ch.CurrentTXV, allowEnd) ||
		!i.decodeNext("index", &i.ch.IdxV, noOpts) ||
		!i.decodeNext("params", i.ch.ParamsV, noOpts) ||
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
