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
	"testing"

	"github.com/stretchr/testify/assert"

	"polycry.pt/poly-go/test"
)

func TestAppRandomizerSet(t *testing.T) {
	test.OnlyOnce(t)

	assert.NotNil(t, appRandomizer, "appRandomizer should be default initialized")
	assert.False(t, isAppRandomizerSet, "isAppRandomizerSet should be defaulted to false")

	old := appRandomizer
	assert.NotPanics(t, func() { SetAppRandomizer(&MockAppRandomizer{}) }, "first SetAppRandomizer() should work")
	assert.True(t, isAppRandomizerSet, "isAppRandomizerSet should be true")
	assert.NotNil(t, appRandomizer, "appRandomizer should not be nil")
	assert.False(t, old == appRandomizer, "appRandomizer should have changed")

	old = appRandomizer
	assert.Panics(t, func() { SetAppRandomizer(&MockAppRandomizer{}) }, "second SetAppRandomizer() should panic")
	assert.True(t, isAppRandomizerSet, "isAppRandomizerSet should be true")
	assert.NotNil(t, appRandomizer, "appRandomizer should not be nil")
	assert.True(t, old == appRandomizer, "appRandomizer should not have changed")
}
