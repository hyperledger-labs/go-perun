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

type (
	// Cache is a message cache.
	Cache struct {
		msgs  []*Envelope
		preds map[*Predicate]struct{}
	}

	// A Predicate defines a message filter.
	Predicate = func(*Envelope) bool

	// A Cacher has the Cache method to enable caching of messages.
	Cacher interface {
		// Cache should enable the caching of messages
		Cache(*Predicate)
	}
)

// MakeCache creates a new cache.
func MakeCache() Cache {
	return Cache{
		preds: make(map[*func(*Envelope) bool]struct{}),
	}
}

// Cache is a message cache. The default value is a valid empty cache.
func (c *Cache) Cache(p *Predicate) {
	c.preds[p] = struct{}{}
}

// Release releases the cache predicate.
func (c *Cache) Release(p *Predicate) {
	delete(c.preds, p)
}

// Put puts the message into the cache if it matches any active predicate.
// If it matches several predicates, it is still only added once to the cache.
func (c *Cache) Put(e *Envelope) bool {
	// we filter the predicates for non-active and lazily remove them
	found := false
	for p := range c.preds {
		found = found || (*p)(e)
	}

	if found {
		c.msgs = append(c.msgs, e)
	}

	return found
}

// Messages retrieves all messages from the cache that match the predicate. They are
// removed from the Cache.
func (c *Cache) Messages(p Predicate) []*Envelope {
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
