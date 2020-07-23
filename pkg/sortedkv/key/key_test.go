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

package key_test

import (
	"testing"

	"perun.network/go-perun/pkg/sortedkv/key"

	"github.com/stretchr/testify/assert"
)

func TestNext(t *testing.T) {
	assert.Equal(t, key.Next(""), "\x00")
	assert.Equal(t, key.Next("a"), "a\x00")
}

func TestIncPrefix(t *testing.T) {
	assert.Equal(t, key.IncPrefix(""), "")
	assert.Equal(t, key.IncPrefix("\x00"), "\x01")
	assert.Equal(t, key.IncPrefix("a"), "b")
	assert.Equal(t, key.IncPrefix("zoo"), "zop")
	assert.Equal(t, key.IncPrefix("\xff"), "")
	assert.Equal(t, key.IncPrefix("\xffa"), "\xffb")
	assert.Equal(t, key.IncPrefix("a\xff"), "b")
	assert.Equal(t, key.IncPrefix("\xff\xff\xff"), "")
}
