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

package client

import (
	"context"
	"sync"

	"perun.network/go-perun/channel"
)

type (
	updateFilter func(ChannelUpdate) bool

	updateAndResponder struct {
		update    ChannelUpdate
		responder *UpdateResponder
	}

	updateInterceptor struct {
		filter   updateFilter
		update   chan updateAndResponder
		response chan struct{}
	}

	updateInterceptors struct {
		entries map[string]*updateInterceptor
		sync.RWMutex
	}
)

func newUpdateInterceptor(filter updateFilter) *updateInterceptor {
	return &updateInterceptor{filter, make(chan updateAndResponder), make(chan struct{})}
}

func (ui *updateInterceptor) HandleUpdate(u ChannelUpdate, r *UpdateResponder) {
	ui.update <- updateAndResponder{u, r}
	<-ui.response
}

func (ui *updateInterceptor) Accept(ctx context.Context) error {
	select {
	case ur := <-ui.update:
		if err := ur.responder.Accept(ctx); err != nil {
			return err
		}
		ui.response <- struct{}{}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func newUpdateInterceptors() *updateInterceptors {
	return &updateInterceptors{entries: make(map[string]*updateInterceptor)}
}

// Register assigns the given update interceptor to the given channel.
func (interceptors *updateInterceptors) Register(id map[int]channel.ID, ui *updateInterceptor) {
	interceptors.Lock()
	defer interceptors.Unlock()
	interceptors.entries[channel.IDKey(id)] = ui
}

// UpdateInterceptor gets the update interceptor for the given channel. The second return
// value indicates whether such an entry could be found.
func (interceptors *updateInterceptors) UpdateInterceptor(id map[int]channel.ID) (*updateInterceptor, bool) {
	interceptors.RLock()
	defer interceptors.RUnlock()
	ui, ok := interceptors.entries[channel.IDKey(id)]
	return ui, ok
}

// Release releases the update interceptor for the given channel.
func (interceptors *updateInterceptors) Release(id map[int]channel.ID) {
	interceptors.Lock()
	defer interceptors.Unlock()
	if ui, ok := interceptors.entries[channel.IDKey(id)]; ok {
		close(ui.response)
	}
	delete(interceptors.entries, channel.IDKey(id))
}

// Filter filters for a matching update interceptor. It returns the first
// matching interceptor. The second return value indicates whether a matching
// interceptor has been found.
func (interceptors *updateInterceptors) Filter(u ChannelUpdate) (*updateInterceptor, bool) {
	interceptors.RLock()
	defer interceptors.RUnlock()
	for _, ui := range interceptors.entries {
		if ui.filter(u) {
			return ui, true
		}
	}
	return nil, false
}
