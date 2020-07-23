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

// Package key of sortedkv provides helper functions to manipulate db keys
package key // import "perun.network/go-perun/pkg/sortedkv/key"

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
