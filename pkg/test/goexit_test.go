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
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckAbort(t *testing.T) {
	abort := CheckAbort(func() {})
	assert.Nil(t, abort)

	abort = CheckAbort(func() { panic(1) })
	require.IsType(t, abort, (*Panic)(nil))
	assert.Equal(t, abort.(*Panic).Value(), 1)

	abort = CheckAbort(func() { panic(nil) })
	require.IsType(t, abort, (*Panic)(nil))
	assert.Nil(t, abort.(*Panic).Value())

	abort = CheckAbort(runtime.Goexit)
	assert.IsType(t, abort, (*Goexit)(nil))
}

func TestCheckGoexit(t *testing.T) {
	assert.True(t, CheckGoexit(runtime.Goexit))
	assert.Panics(t, func() { CheckGoexit(func() { panic("") }) })
	didPanic, pval := CheckPanic(func() { CheckGoexit(func() { panic(nil) }) })
	assert.True(t, didPanic)
	assert.Nil(t, pval)
	assert.False(t, CheckGoexit(func() {}))
}
