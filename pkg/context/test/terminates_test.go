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

package test

import (
	"context"
	"testing"
	"time"

	"perun.network/go-perun/pkg/test"
)

const timeout = 200 * time.Millisecond

func TestAssertTerminatesCtx(t *testing.T) {
	t.Run("error case", func(t *testing.T) {
		test.AssertError(t, func(t test.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			AssertTerminatesCtx(ctx, t, func() {})
		})
	})

	t.Run("success case", func(t *testing.T) {
		AssertTerminatesCtx(context.Background(), t, func() {})
	})
}

func TestAssertNotTerminatesCtx(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		AssertNotTerminatesCtx(ctx, t, func() {})
	})

	t.Run("error case", func(t *testing.T) {
		test.AssertError(t, func(t test.T) {
			AssertNotTerminatesCtx(context.Background(), t, func() {})
		})
	})
}

func TestAssertTerminates(t *testing.T) {
	t.Run("error case", func(t *testing.T) {
		test.AssertError(t, func(t test.T) {
			AssertTerminates(t, -1, func() {})
		})
	})

	t.Run("success case", func(t *testing.T) {
		AssertTerminates(t, timeout, func() {})
	})
}

func TestAssertNotTerminates(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		AssertNotTerminates(t, -1, func() {})
	})

	t.Run("error case", func(t *testing.T) {
		test.AssertError(t, func(t test.T) {
			AssertNotTerminates(t, timeout, func() {})
		})
	})
}
