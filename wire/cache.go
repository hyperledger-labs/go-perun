// Copyright 2019 - See NOTICE file for copyright holders.
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

package wire

import "context"

type (
	// Cache is a message cache. The default value is a valid empty cache.
	Cache struct {
		msgs  []*Envelope
		preds []ctxPredicate
	}

	// A Predicate defines a message filter.
	Predicate = func(*Envelope) bool

	ctxPredicate struct {
		ctx context.Context
		p   Predicate
	}

	// A Cacher has the Cache method to enable caching of messages.
	Cacher interface {
		// Cache should enable the caching of messages
		Cache(context.Context, Predicate)
	}
)

// Cache is a message cache. The default value is a valid empty cache.
func (c *Cache) Cache(ctx context.Context, p Predicate) {
	c.preds = append(c.preds, ctxPredicate{ctx, p})
}

// Put puts the message into the cache if it matches any active predicate.
// If it matches several predicates, it is still only added once to the cache.
func (c *Cache) Put(e *Envelope) bool {
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

		any = any || p.p(e)
	}

	if any {
		c.msgs = append(c.msgs, e)
	}

	c.preds = preds
	return any
}

// Get retrieves all messages from the cache that match the predicate. They are
// removed from the Cache.
func (c *Cache) Get(p Predicate) []*Envelope {
	msgs := c.msgs[:0]
	// Usually, Get is called with the assumption to match at least one message
	matches := make([]*Envelope, 0, 1)
	for _, m := range c.msgs {
		if p(m) {
			matches = append(matches, m)
		} else {
			msgs = append(msgs, m)
		}
	}
	c.msgs = msgs
	return matches
}

// Flush empties the message cache and removes all predicates.
func (c *Cache) Flush() {
	c.msgs = nil
	c.preds = nil
}

// Size returns the number of messages held in the message cache.
func (c *Cache) Size() int {
	return len(c.msgs)
}
