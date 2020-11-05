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

package sync_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/context/test"
	"perun.network/go-perun/pkg/sync"
)

func TestWaitGoup(t *testing.T) {
	var wg sync.WaitGroup

	// Empty waitgroups have a counter of 0, and should immediately return.
	test.AssertTerminatesQuickly(t, wg.Wait)

	t.Run("negative counter", func(t *testing.T) {
		assert.Panics(t, func() { wg.Add(-5) })
		assert.Panics(t, wg.Done)
	})

	wg.Add(1)
	test.AssertNotTerminatesQuickly(t, wg.Wait)
	wg.Done()
	test.AssertTerminatesQuickly(t, wg.Wait)
	test.AssertTerminatesQuickly(t, wg.Wait)

	const N = 5
	wg.Add(N)
	test.AssertNotTerminatesQuickly(t, wg.Wait)

	for i := 0; i < N-1; i++ {
		wg.Done()
		test.AssertNotTerminatesQuickly(t, wg.Wait)
	}
	wg.Done()
	test.AssertTerminatesQuickly(t, wg.Wait)
}
