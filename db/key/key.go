// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

// Package key of db provides helper functions to manipulate db keys
package key // import "perun.network/go-perun/db/key"

// Next returns the key with a zero byte appended, which is the next key in the
// lexicographical order of strings
// Useful for NewIteratorWithRange if the end should be included.
func Next(key string) string {
	return key + "\x00"
}

// IncPrefix increments a prefix string, such that
// for all prefix,suffix: prefix+suffix < IncrementPrefix(prefix).
// If the empty string or a string where all bits are 1 is passed, the empty string
// is returned, indicating no upper limit.
// This is useful for string range calculations
func IncPrefix(key string) string {
	keyb := []byte(key)
	overflows := 0
	for i := len(keyb) - 1; i >= 0; i-- {
		// Increment current byte, stop if it doesn't overflow
		keyb[i]++
		if keyb[i] > 0 {
			break
		} else {
			overflows++
		}
		// Character overflown, proceed to next or return "" if last
		if i == 0 {
			return ""
		}
	}
	return string(keyb[:len(keyb)-overflows])
}
