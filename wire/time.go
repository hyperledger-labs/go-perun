// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"time"
)

// Time is a serializable network timestamp.
// It is a 64-bit unix timestamp, in nanoseconds.
type Time struct {
	Int64
}

// Time converts a wire Time into a system time.
func (t Time) Time() time.Time {
	return time.Unix(0, int64(t.Int64))
}

// FromTime creates a wire Time from a system time.
func FromTime(time time.Time) Time {
	return Time{Int64(time.UnixNano())}
}

// Now creates a wire Time representing the current moment.
func Now() Time {
	return Time{Int64(time.Now().UnixNano())}
}
