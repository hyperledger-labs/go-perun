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

package wire_test

import (
	"perun.network/go-perun/wallet"
	"testing"

	. "perun.network/go-perun/wire"
	"perun.network/go-perun/wire/test"
)

func TestLocalBus(t *testing.T) {
	bus := NewLocalBus()
	test.GenericBusTest(t, func(map[wallet.BackendID]Account) (Bus, Bus) {
		return bus, bus
	}, 16, 10)
}
