// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import "context"

type (
	// Cache is a message cache. The default value is a valid empty cache.
	Cache struct {
		msgs  []MsgAnnex
		preds []ctxPredicate
	}

	// MsgAnnex is a tuple of a message together with some arbitrary additional
	// data (Annex)
	MsgAnnex struct {
		Msg   Msg
		Annex interface{}
	}

	// A Predicate defines a message filter.
	Predicate = func(Msg) bool

	ctxPredicate struct {
		ctx context.Context
		p   Predicate
	}
)

// Cache is a message cache. The default value is a valid empty cache.
func (c *Cache) Cache(ctx context.Context, p Predicate) {
	c.preds = append(c.preds, ctxPredicate{ctx, p})
}

// Put puts the message into the cache if it matches any active prediacte.
// If it matches several predicates, it is still only added once to the cache.
func (c *Cache) Put(m Msg, a interface{}) bool {
	// we filter the predicates for non-active and lazily remove them
	preds := c.preds[:0]
	any := false
	for _, p := range c.preds {
		select {
		case <-p.ctx.Done():
			continue // skip done predicate
		default:
			preds = append(preds, p)
		}

		any = any || p.p(m)
	}

	if any {
		c.msgs = append(c.msgs, MsgAnnex{m, a})
	}

	c.preds = preds
	return any
}

// Get retrieves all messages from the cache that match the predicate. They are
// removed from the Cache.
func (c *Cache) Get(p Predicate) []MsgAnnex {
	msgs := c.msgs[:0]
	// Usually, Get is called with the assumption to match at least one message
	matches := make([]MsgAnnex, 0, 1)
	for _, m := range c.msgs {
		if p(m.Msg) {
			matches = append(matches, m)
		} else {
			msgs = append(msgs, m)
		}
	}
	c.msgs = msgs
	return matches
}

// Flush empties the message cache. Predicates are retained.
func (c *Cache) Flush() {
	c.msgs = nil
	c.preds = nil
}

// Size returns the number of messages held in the message cache.
func (c *Cache) Size() int {
	return len(c.msgs)
}
