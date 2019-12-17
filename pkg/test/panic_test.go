// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

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
