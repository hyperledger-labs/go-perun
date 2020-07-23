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
	"runtime"
	"sync"
)

// Skipper is a subset of the testing.T functionality needed by OnlyOnce().
type Skipper interface {
	SkipNow()
}

var executedTests = make(map[string]struct{})
var executedTestsMutex sync.Mutex

// OnlyOnce records a test case and skips it if it already executed once.
// Test case identification is done by observing the stack. Calls SkipNow() on
// tests that have already been executed. OnlyOnce() has to be called directly
// from the test's function, as its first action.
func OnlyOnce(t Skipper) {
	pc, _, _, _ := runtime.Caller(1)
	name := runtime.FuncForPC(pc).Name()

	executedTestsMutex.Lock()
	defer executedTestsMutex.Unlock()

	if _, executed := executedTests[name]; executed {
		t.SkipNow()
	}
	executedTests[name] = struct{}{}
}
