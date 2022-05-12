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

package channel

import (
	"github.com/pkg/errors"

	"perun.network/go-perun/log"
)

type (
	noApp  struct{}
	noData struct{}
)

// NoApp returns an empty app that contains no logic.
func NoApp() App { return noApp{} }

// IsNoApp checks whether an app is a NoApp.
func IsNoApp(a App) bool {
	_, ok := a.(noApp)
	return ok
}

var _ StateApp = noApp{}

// Def panics and should not be called.
func (noApp) Def() AppID {
	log.Panic("must not call Def() on NoApp")
	return nil // needed to keep the compiler happy.
}

// NewData returns a new instance of data specific to NoApp, intialized to its
// zero value.
//
// This should be used for unmarshalling the data from its binary
// representation.
func (noApp) NewData() Data {
	return NoData()
}

// ValidTransition allows all transitions.
func (noApp) ValidTransition(*Params, *State, *State, Index) error {
	return nil
}

// ValidInit expects the state to have NoData.
func (noApp) ValidInit(_ *Params, s *State) error {
	if !IsNoData(s.Data) {
		return errors.Errorf("State must have NoData, has %T", s.Data)
	}
	return nil
}

// NoData creates an empty app data value.
func NoData() Data { return &noData{} }

// IsNoData returns whether an app data is NoData.
func IsNoData(d Data) bool {
	_, ok := d.(*noData)
	return ok
}

// MarshalBinary does nothing and always returns an empty byte array and nil.
func (noData) MarshalBinary() ([]byte, error) { return []byte{}, nil }

// UnmarshalBinary does nothing and always returns nil.
func (*noData) UnmarshalBinary(_ []byte) error { return nil }

// Clone returns NoData().
func (noData) Clone() Data { return NoData() }
