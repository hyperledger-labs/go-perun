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

package test

import (
	"math/rand"

	"perun.network/go-perun/channel"
)

// The AppRandomizer interface provides functionality to create RandomApps and
// RandomData which is useful for testing.
type AppRandomizer interface {
	NewRandomApp(*rand.Rand) channel.App
	NewRandomData(*rand.Rand) channel.Data
}

var appRandomizer AppRandomizer = &MockAppRandomizer{}

// isAppRandomizerSet whether the appRandomizer was already set with `SetAppRandomizer`
var isAppRandomizerSet bool

// SetAppRandomizer sets the global appRandomizer.
func SetAppRandomizer(r AppRandomizer) {
	if isAppRandomizerSet {
		panic("app randomizer already set")
	}
	isAppRandomizerSet = true
	appRandomizer = r
}

// NewRandomApp creates a new random channel.App.
func NewRandomApp(rng *rand.Rand, opts ...RandomOpt) channel.App {
	opt := mergeRandomOpts(opts...)

	if app := opt.App(); app != nil {
		return app
	}
	if def := opt.AppDef(); def != nil {
		app, _ := channel.AppFromDefinition(def)
		return app
	}
	// WithAppDef does not set the app in the options
	app := appRandomizer.NewRandomApp(rng)
	updateOpts(opts, WithApp(app))
	return app
}

// NewRandomData creates new random data for an app.
func NewRandomData(rng *rand.Rand, opts ...RandomOpt) channel.Data {
	opt := mergeRandomOpts(opts...)

	if data := opt.AppData(); data != nil {
		return data
	}

	data := appRandomizer.NewRandomData(rng)
	updateOpts(opts, WithAppData(data))
	return data
}
