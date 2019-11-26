// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

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
