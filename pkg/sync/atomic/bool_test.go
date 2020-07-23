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

package atomic_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/sync/atomic"
)

func TestBool(t *testing.T) {
	assert := assert.New(t)

	var b atomic.Bool
	assert.False(b.IsSet())
	b.Set()
	assert.True(b.IsSet())
	assert.False(b.TrySet())
	assert.True(b.IsSet())

	b.Unset()
	assert.False(b.IsSet())
	assert.True(b.TrySet())
	assert.True(b.IsSet())
	assert.True(b.TryUnset())
	assert.False(b.TryUnset())
	assert.False(b.IsSet())
}
