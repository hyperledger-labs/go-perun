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

// The AppRandomizer interface provides functionality for creating random
// data and apps which is useful for testing.
type AppRandomizer interface {
	NewRandomApp(*rand.Rand) channel.App
	NewRandomData(*rand.Rand) channel.Data
}

var appRandomizer AppRandomizer = &MockAppRandomizer{}

// isAppRandomizerSet tracks whether the AppRandomizer was already set
// with `SetAppRandomizer`.
var isAppRandomizerSet bool

// SetAppRandomizer sets the global `appRandomizer`.
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
		app, _ := channel.Resolve(def)
		return app
	}
	// WithAppDef does not set the app in the options
	app := opt.AppRandomizer().NewRandomApp(rng)
	channel.RegisterApp(app)
	updateOpts(opts, WithApp(app))
	return app
}

// NewRandomData creates new random data for an app.
func NewRandomData(rng *rand.Rand, opts ...RandomOpt) channel.Data {
	opt := mergeRandomOpts(opts...)

	if data := opt.AppData(); data != nil {
		return data
	}

	data := opt.AppRandomizer().NewRandomData(rng)
	updateOpts(opts, WithAppData(data))
	return data
}

// NewRandomAppAndData creates a new random channel.App and new random channel.Data.
func NewRandomAppAndData(rng *rand.Rand, opts ...RandomOpt) (channel.App, channel.Data) {
	opt := mergeRandomOpts(opts...)
	return NewRandomApp(rng, opt), NewRandomData(rng, opt)
}

// NewRandomAppIDFunc is an app identifier randomizer function.
type NewRandomAppIDFunc = func(*rand.Rand) channel.AppID

var newRandomAppID NewRandomAppIDFunc

func SetNewRandomAppID(f NewRandomAppIDFunc) {
	newRandomAppID = f
}

// NewRandomAppID creates a new random channel.AppID.
func NewRandomAppID(rng *rand.Rand) channel.AppID {
	return newRandomAppID(rng)
}
