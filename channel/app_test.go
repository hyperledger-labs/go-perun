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

package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/test"
)

func TestAppBackendSet(t *testing.T) {
	test.OnlyOnce(t)

	assert.NotNil(t, appBackend, "appBackend should be default initialized")
	assert.False(t, isAppBackendSet, "isAppBackendSet should be defaulted to false")

	old := appBackend
	assert.NotPanics(t, func() { SetAppBackend(&MockAppBackend{}) }, "first SetAppBackend() should work")
	assert.True(t, isAppBackendSet, "isAppBackendSet should be true")
	assert.NotNil(t, appBackend, "appBackend should not be nil")
	assert.False(t, old == appBackend, "appBackend should have changed")

	old = appBackend
	assert.Panics(t, func() { SetAppBackend(&MockAppBackend{}) }, "second SetAppBackend() should panic")
	assert.True(t, isAppBackendSet, "isAppBackendSet should be true")
	assert.NotNil(t, appBackend, "appBackend should not be nil")
	assert.True(t, old == appBackend, "appBackend should not have changed")
}
