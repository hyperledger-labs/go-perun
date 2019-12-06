// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/wire/msg"
)

// Consumer consumes messages fed into it via Put().
type Consumer interface {
	// The producer calls OnClose() to unregister the Consumer after it is
	// closed.
	sync.OnCloser
	// Put is called by the emitter when relaying a message.
	Put(*Peer, msg.Msg)
}
