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
)

func TestConcurrentT_Wait(t *testing.T) {
	t.Run("failed stage", func(t *testing.T) {
		ct := NewConcurrent(t)
		s := ct.spawnStage("stage", 1)
		s.failed.Set()
		s.pass()

		assert.True(t, CheckGoexit(func() { ct.Wait("stage") }),
			"Waiting for a failed stage must call runtime.Goexit.")
	})
}

func TestStage_FailNow(t *testing.T) {
	t.Run("first fail", func(t *testing.T) {
		AssertFatal(t, func(t T) {
			ct := NewConcurrent(t)
			s := ct.spawnStage("stage", 1)
			s.FailNow()
		})
	})

	t.Run("second fail", func(t *testing.T) {
		ct := NewConcurrent(nil)
		ct.failed = true
		s := ct.spawnStage("stage", 1)
		assert.True(t, CheckGoexit(s.FailNow))
	})
}
