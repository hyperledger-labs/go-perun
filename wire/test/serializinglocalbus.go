// Copyright 2021 - See NOTICE file for copyright holders.
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

package test

import (
	"bytes"
	"context"

	"perun.network/go-perun/wire"
	// nolint: blank-imports  // allow blank import package that is not main or test.
	_ "perun.network/go-perun/wire/perunio/serializer"
)

// SerializingLocalBus is a local bus that also serializes messages for testing.
type SerializingLocalBus struct {
	*wire.LocalBus
}

// NewSerializingLocalBus creates a new serializing local bus.
func NewSerializingLocalBus() *SerializingLocalBus {
	return &SerializingLocalBus{
		LocalBus: wire.NewLocalBus(),
	}
}

// Publish publishes the message on the bus.
func (b *SerializingLocalBus) Publish(ctx context.Context, e *wire.Envelope) (err error) {
	var buf bytes.Buffer
	err = wire.EncodeEnvelope(&buf, e)
	if err != nil {
		return
	}

	_e, err := wire.DecodeEnvelope(&buf)
	if err != nil {
		return
	}
	return b.LocalBus.Publish(ctx, _e)
}
