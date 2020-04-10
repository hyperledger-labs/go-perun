// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package persistence

import (
	"context"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/peer"
)

// NonPersister is a Persister that doesn't to anything. All its methods return
// nil.
var NonPersister Persister = nonPersister{}

type nonPersister struct{}

func (nonPersister) ChannelCreated(context.Context, Source, []peer.Address) error { return nil }
func (nonPersister) ChannelRemoved(context.Context, channel.ID) error             { return nil }
func (nonPersister) Staged(context.Context, Source) error                         { return nil }
func (nonPersister) SigAdded(context.Context, Source, channel.Index) error        { return nil }
func (nonPersister) Enabled(context.Context, Source) error                        { return nil }
func (nonPersister) PhaseChanged(context.Context, Source) error                   { return nil }
func (nonPersister) Close() error                                                 { return nil }
