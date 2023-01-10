// Copyright 2019 - See NOTICE file for copyright holders.
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

package payment

import (
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

// Resolver is the payment app resolver.
type Resolver struct{}

// Resolve returns a payment app with the given definition.
func (b *Resolver) Resolve(def wallet.Address) (channel.App, error) {
	return &App{def}, nil
}
