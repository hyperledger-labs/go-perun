// Copyright 2021 - See NOTICE file for copyright holders.
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

package test_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/backend/ethereum/channel/test"
	pkgtest "polycry.pt/poly-go/test"
)

// TestSimBackend_AutoMine tests that SimulatedBackend mines blocks after
// `StartMining` and before `StopMining` was called.
func TestSimBackend_AutoMine(t *testing.T) {
	sb := test.NewSimulatedBackend()

	// Start mining with 10 blocks/second.
	sb.StartMining(100 * time.Millisecond)

	// Assert that it produced at least 5 blocks within the next second.
	pkgtest.Within1s.Eventually(t, func(t pkgtest.T) {
		head, err := sb.HeaderByNumber(context.Background(), nil)
		require.NoError(t, err)
		assert.Greater(t, head.Number.Uint64(), uint64(4))
	})

	// Stop mining.
	sb.StopMining()

	// Wait half a second.
	time.Sleep(500 * time.Millisecond)

	// Assert that it is not producing blocks anymore.
	head1, err := sb.HeaderByNumber(context.Background(), nil)
	require.NoError(t, err)
	time.Sleep(500 * time.Millisecond)
	head2, err := sb.HeaderByNumber(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, head1.Number.Uint64(), head2.Number.Uint64())
}
