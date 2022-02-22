// Copyright 2022 - See NOTICE file for copyright holders.
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

package wire_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "perun.network/go-perun/wire"
	"perun.network/go-perun/wire/test"
)

func TestHybridBus_New(t *testing.T) {
	require.Panics(t, func() { NewHybridBus() })
	require.Panics(t, func() { NewHybridBus(nil) })
	require.NotPanics(t, func() { NewHybridBus(NewLocalBus()) })
	require.Panics(t, func() { NewHybridBus(NewLocalBus(), nil) })
	require.NotPanics(t, func() { NewHybridBus(NewLocalBus(), NewLocalBus()) })
}

func TestHybridBus(t *testing.T) {
	nBuses := 5
	buses := make([]Bus, nBuses)
	for i := range buses {
		buses[i] = NewLocalBus()
	}

	hybridBus := NewHybridBus(buses...)

	i := 0
	test.GenericBusTest(t, func(Account) (pub Bus, sub Bus) {
		i++
		// Split the clients evenly among the sub-buses, and let them publish
		// over the hybrid bus.
		return hybridBus, buses[i%nBuses]
	}, 16, 10)
}
