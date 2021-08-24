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

package client_test

import (
	"github.com/sirupsen/logrus"

	"perun.network/go-perun/apps/payment"
	"perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/channel/test"
	plogrus "perun.network/go-perun/log/logrus"
)

func init() {
	plogrus.Set(logrus.WarnLevel, &logrus.TextFormatter{ForceColors: true})

	// Eth client tests use the payment app for now...
	// TODO: This has to be set to the deployed app contract (or counterfactual
	// address of it) when we start using it in tests.
	test.SetAppRandomizer(new(payment.Randomizer))
	// Fix the finality depth for testing.
	channel.TxFinalityDepth = 3
}
