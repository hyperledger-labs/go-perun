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
	"github.com/stretchr/testify/require"
)

func TestDialerList_insert(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	var l dialerList

	d := &Dialer{}
	l.insert(d)
	require.Len(l.entries, 1)
	assert.Same(d, l.entries[0])

	d2 := &Dialer{}
	l.insert(d2)
	require.Len(l.entries, 2)
	assert.Same(d2, l.entries[1])
}

func TestDialerList_erase(t *testing.T) {
	assert := assert.New(t)
	var l dialerList

	assert.Error(l.erase(&Dialer{}))
	d := &Dialer{}
	l.insert(d)
	assert.NoError(l.erase(d))
	assert.Len(l.entries, 0)
	assert.Error(l.erase(d))
}

func TestDialerList_clear(t *testing.T) {
	assert := assert.New(t)
	var l dialerList

	d := &Dialer{}
	l.insert(d)
	assert.Equal(l.clear(), []*Dialer{d})
	assert.Len(l.entries, 0)
}
