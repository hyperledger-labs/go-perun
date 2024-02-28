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

package channel_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestAppShouldEqual(t *testing.T) {
	rng := pkgtest.Prng(t)
	app1 := test.NewRandomApp(rng)
	app2 := test.NewRandomApp(rng)
	napp := channel.NoApp()

	require.EqualError(t, channel.AppShouldEqual(app1, app2), "different App definitions")
	require.EqualError(t, channel.AppShouldEqual(app2, app1), "different App definitions")
	require.NoError(t, channel.AppShouldEqual(app1, app1))
	require.EqualError(t, channel.AppShouldEqual(app1, napp), "(non-)nil App definitions")
	require.EqualError(t, channel.AppShouldEqual(napp, app1), "(non-)nil App definitions")
	require.NoError(t, channel.AppShouldEqual(napp, napp))
}
