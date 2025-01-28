// Copyright 2025 - See NOTICE file for copyright holders.
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

package local

import (
	"sync"

	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
)

type (
	registry struct {
		mtx sync.Mutex
		chs map[string]*ch
	}
)

func newRegistry() *registry {
	return &registry{
		chs: make(map[string]*ch),
	}
}

// addIfSucceeds adds the channel to the registry, if it is not already present
// in the registry and if the "chInitializer" does not return an error.
func (r *registry) addIfSucceeds(id map[wallet.BackendID]channel.ID, chInitializer chInitializer) (*ch, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if _, ok := r.chs[channel.IDKey(id)]; ok {
		return nil, errors.New("already watching for this channel")
	}

	ch, err := chInitializer()
	if err != nil {
		return nil, err
	}

	r.chs[channel.IDKey(ch.id)] = ch
	return ch, nil
}

// retrieve retrieves the channel from registry.
func (r *registry) retrieve(id map[wallet.BackendID]channel.ID) (*ch, bool) {
	r.mtx.Lock()
	ch, ok := r.chs[channel.IDKey(id)]
	r.mtx.Unlock()
	return ch, ok
}

// remove removes the channel from registry, if it is present.
// It does not do any validation on the channel to be removed.
func (r *registry) remove(id map[wallet.BackendID]channel.ID) {
	r.mtx.Lock()
	delete(r.chs, channel.IDKey(id))
	r.mtx.Unlock()
}
