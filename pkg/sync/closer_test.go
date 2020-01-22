// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package sync_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/pkg/test"
)

const timeout = 100 * time.Millisecond

func TestCloser_Closed(t *testing.T) {
	t.Parallel()
	var c sync.Closer

	assert.NotNil(t, c.Closed())
	select {
	case _, ok := <-c.Closed():
		t.Fatalf("Closed() should not yield a value, ok = %t", ok)
	default:
	}

	require.NoError(t, c.Close())

	test.AssertTerminates(t, timeout, func() {
		_, ok := <-c.Closed()
		assert.False(t, ok)
	})
}
