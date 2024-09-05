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

package wire

import (
	"context"
	"perun.network/go-perun/wallet"
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
)

type localBusReceiver struct {
	recv   Consumer
	exists chan struct{}
}

var _ Bus = (*LocalBus)(nil)

// LocalBus is a bus that only sends message in the same process.
type LocalBus struct {
	mutex sync.RWMutex
	recvs map[AddrKey]*localBusReceiver
}

// NewLocalBus creates a new local bus, which only targets receivers that lie
// within the same process.
func NewLocalBus() *LocalBus {
	return &LocalBus{recvs: make(map[AddrKey]*localBusReceiver)}
}

// Publish implements wire.Bus.Publish. It returns only once the recipient
// received the message or the context times out.
func (h *LocalBus) Publish(ctx context.Context, e *Envelope) error {
	recv := h.ensureRecv(e.Recipient)
	select {
	case <-recv.exists:
		recv.recv.Put(e)
		return nil
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "publishing message")
	}
}

// SubscribeClient implements wire.Bus.SubscribeClient. There can only be one
// subscription per receiver address.
// When the Consumer closes, its subscription is removed.
func (h *LocalBus) SubscribeClient(c Consumer, receiver map[wallet.BackendID]Address) error {
	recv := h.ensureRecv(receiver)
	recv.recv = c
	close(recv.exists)

	c.OnCloseAlways(func() {
		h.mutex.Lock()
		defer h.mutex.Unlock()
		delete(h.recvs, Keys(receiver))
		log.WithField("id", receiver).Debug("Client unsubscribed.")
	})

	log.WithField("id", receiver).Debug("Client subscribed.")
	return nil
}

// ensureRecv ensures that there is an entry for a recipient address in the
// bus' receiver map, and returns it. If it creates a new receiver, it is only
// a placeholder until a subscription appears.
func (h *LocalBus) ensureRecv(a map[wallet.BackendID]Address) *localBusReceiver {
	key := Keys(a)
	// First, we only use a read lock, hoping that the receiver already exists.
	h.mutex.RLock()
	recv, ok := h.recvs[key]
	h.mutex.RUnlock()

	if ok {
		return recv
	}

	// If not, we have to insert one, so we need exclusive an lock.
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// We need to re-check, because between the RUnlock() and Lock(), it could
	// have been added by another goroutine already.
	recv, ok = h.recvs[key]
	if ok {
		return recv
	}
	// Insert and return the new entry.
	recv = &localBusReceiver{
		recv:   nil,
		exists: make(chan struct{}),
	}
	h.recvs[key] = recv
	return recv
}
