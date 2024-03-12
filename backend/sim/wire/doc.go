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

// Package wire is used for internal tests in the packages channel, wire, and client.
// Note that Account.Sign and Address.Verify are mock methods.
// We use the backend/wire/sim mock implementation for testing other go-perun functionalities.
// Our default wire.Account and wire.Address implementations can be found in wire/net/simple and are used for our applications.
package wire // import "perun.network/go-perun/backend/sim/wire"
