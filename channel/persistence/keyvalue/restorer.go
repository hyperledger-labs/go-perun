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
	"perun.network/go-perun/pkg/sortedkv"
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
		if err = wire.Decode(buf, &id); err != nil {
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

// Next advances the iterator and returns whether there is another channel.
func (i *ChannelIterator) Next(context.Context) bool {
	i.ch = &persistence.Channel{ParamsV: &channel.Params{}}

	return i.decodeNext("current", &i.ch.CurrentTXV, true) &&
		i.decodeNext("index", &i.ch.IdxV, false) &&
		i.decodeNext("params", i.ch.ParamsV, false) &&
		i.decodeNext("phase", &i.ch.PhaseV, false) &&
		i.decodeNext("staging", &i.ch.StagingTXV, false)
}

// decodeNext reduces code duplication for decoding a value from an iterator.
func (i *ChannelIterator) decodeNext(key string, v interface{}, allowEnd bool) bool {
	for !i.its[0].Next() {
		if !allowEnd {
			i.err = errors.WithMessage(i.its[0].Close(), "expected key "+key)
			if i.err == nil {
				i.err = errors.New("expected key " + key)
			}
			return false
		}
		// Move on to the next iterator.
		i.its = i.its[1:]
		if len(i.its) == 0 {
			return false
		}
	}

	buf := bytes.NewBuffer(i.its[0].ValueBytes())
	i.err = errors.WithMessage(wire.Decode(buf, v), "decoding "+key)
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
