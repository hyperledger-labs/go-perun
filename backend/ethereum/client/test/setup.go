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
	"time"

	ethctest "perun.network/go-perun/backend/ethereum/channel/test"
	ethwtest "perun.network/go-perun/backend/ethereum/wallet/test"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/watcher/local"
	"perun.network/go-perun/wire"
)

const (
	DefaultTimeout = 5 * time.Second
	BlockInterval  = 100 * time.Millisecond
)

func MakeRoleSetups(s *ethctest.Setup, names [2]string) (setup [2]clienttest.RoleSetup) {
	bus := wire.NewLocalBus()
	for i := 0; i < len(setup); i++ {
		watcher, err := local.NewWatcher(s.Adjs[i])
		if err != nil {
			panic("Error initializing watcher: " + err.Error())
		}
		setup[i] = clienttest.RoleSetup{
			Name:              names[i],
			Identity:          s.Accs[i],
			Bus:               bus,
			Funder:            s.Funders[i],
			Adjudicator:       s.Adjs[i],
			Watcher:           watcher,
			Wallet:            ethwtest.NewTmpWallet(),
			Timeout:           DefaultTimeout,
			ChallengeDuration: 60 * uint64(time.Second/BlockInterval), // Scaled due to simbackend automining progressing faster than real time.
		}
	}
	return
}
