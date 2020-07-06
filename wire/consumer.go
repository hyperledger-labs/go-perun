// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"perun.network/go-perun/pkg/sync"
)

// Consumer consumes messages fed into it via Put().
type Consumer interface {
	// The producer calls OnClose() to unregister the Consumer after it is
	// closed.
	sync.OnCloser
	// Put is called by the emitter when relaying a message.
	Put(*Peer, Msg)
}
