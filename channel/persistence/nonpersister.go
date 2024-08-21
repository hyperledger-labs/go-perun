// Copyright 2020 - See NOTICE file for copyright holders.
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

package persistence

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire"
)

// NonPersistRestorer is a PersistRestorer that doesn't do anything. All
// Persistence methods return nil and all Restorer methods return an empty
// iterator.
var NonPersistRestorer PersistRestorer = nonPersistRestorer{}

type nonPersistRestorer struct{}

// Persister implementation

func (nonPersistRestorer) ChannelCreated(context.Context, channel.Source, []map[int]wire.Address, *channel.ID) error {
	return nil
}
func (nonPersistRestorer) ChannelRemoved(context.Context, channel.ID) error              { return nil }
func (nonPersistRestorer) Staged(context.Context, channel.Source) error                  { return nil }
func (nonPersistRestorer) SigAdded(context.Context, channel.Source, channel.Index) error { return nil }
func (nonPersistRestorer) Enabled(context.Context, channel.Source) error                 { return nil }
func (nonPersistRestorer) PhaseChanged(context.Context, channel.Source) error            { return nil }
func (nonPersistRestorer) Close() error                                                  { return nil }

// Restorer implementation

func (nonPersistRestorer) ActivePeers(context.Context) ([]map[int]wire.Address, error) {
	return nil, nil
}

func (nonPersistRestorer) RestoreAll() (ChannelIterator, error) {
	return emptyChanIterator{}, nil
}

func (nonPersistRestorer) RestorePeer(map[int]wire.Address) (ChannelIterator, error) {
	return emptyChanIterator{}, nil
}

func (nonPersistRestorer) RestoreChannel(context.Context, channel.ID) (*Channel, error) {
	return nil, errors.New("channel not found")
}

type emptyChanIterator struct{}

func (emptyChanIterator) Next(context.Context) bool { return false }
func (emptyChanIterator) Channel() *Channel         { return nil }
func (emptyChanIterator) Close() error              { return nil }
