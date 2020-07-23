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

import "testing"

func TestCheckPanic(t *testing.T) {
	// Test whether panic calls are properly detected and whether the supplied
	// value is also properly recorded.
	if p, v := CheckPanic(func() { panic("panicvalue") }); !p || v != "panicvalue" {
		t.Error("Failed to detect panic!")
	}

	// Test whether panic(nil) calls are detected.
	if p, v := CheckPanic(func() { panic(nil) }); !p || v != nil {
		t.Error("Failed to detect panic(nil)!")
	}

	// Test whether the absence of panic calls is properly detected.
	if p, v := CheckPanic(func() {}); p || v != nil {
		t.Error("False positive panic detection!")
	}
}
