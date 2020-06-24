// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package keyvalue

import (
	"bytes"
	"context"
	"io"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/peer"
	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/pkg/sortedkv"
	"perun.network/go-perun/wallet"
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
func (r *PersistRestorer) ActivePeers(context.Context) ([]peer.Address, error) {
	ps := make([]peer.Address, 0, len(r.cache.peerChannels))
	for peerstr := range r.cache.peerChannels {
		addr, err := peer.DecodeAddress(bytes.NewReader([]byte(peerstr)))
		if err != nil {
			return nil, errors.WithMessagef(err, "decoding peer address (%x)", []byte(peerstr))
		}
		ps = append(ps, addr)
	}
	return ps, nil
}

// RestoreAll should return an iterator over all persisted channels.
func (r *PersistRestorer) RestoreAll() (persistence.ChannelIterator, error) {
	return &ChannelIterator{
		restorer: r,
		its:      []sortedkv.Iterator{sortedkv.NewTable(r.db, "Chan:").NewIterator()},
	}, nil
}

// RestorePeer should return an iterator over all persisted channels which
// the given peer is a part of.
func (r *PersistRestorer) RestorePeer(addr peer.Address) (persistence.ChannelIterator, error) {
	it := &ChannelIterator{restorer: r}
	chandb := sortedkv.NewTable(r.db, "Chan:")

	chs := r.cache.peerChannels[string(addr.Bytes())]
	it.its = make([]sortedkv.Iterator, len(chs))
	i := 0
	for ch := range chs {
		it.its[i] = chandb.NewIteratorWithPrefix(string(ch[:]))
		i++
	}

	return it, nil
}

// RestoreChannel restores a single channel.
func (r *PersistRestorer) RestoreChannel(ctx context.Context, id channel.ID) (*persistence.Channel, error) {
	it := &ChannelIterator{restorer: r}
	chandb := sortedkv.NewTable(r.db, "Chan:")

	it.its = append(it.its, chandb.NewIteratorWithPrefix(string(id[:])))

	if it.Next(ctx) {
		return it.Channel(), it.Close()
	}
	return nil, it.Close()
}

// readAllPeers reads all peer entries from the database and populates the
// restorer's channel cache.
func (r *PersistRestorer) readAllPeers() (err error) {
	it := sortedkv.NewTable(r.db, "Peer:").NewIterator()
	defer it.Close()

	for it.Next() {
		buf := bytes.NewBufferString(it.Key())

		var addr peer.Address
		if addr, err = peer.DecodeAddress(buf); err != nil {
			return errors.WithMessage(err, "decode peer address")
		}

		if err = eatExpect(buf, ":channel:"); err != nil {
			return errors.WithMessagef(err, "key: %s", it.Key())
		}

		var id channel.ID
		if err = perunio.Decode(buf, &id); err != nil {
			return errors.WithMessage(err, "decode channel id")
		}

		r.cache.addPeerChannel(addr, id)
	}

	return errors.WithMessage(it.Close(), "iterator")
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
