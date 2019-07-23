// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import "testing"

func TestCheckPanic(t *testing.T) {
	// Test whether panic calls are properly detected and whether the supplied
	// value is also properly recorded.
	if "panicvalue" != CheckPanic(func() { panic("panicvalue") }) {
		t.Error("Failed to detect panic!")
	}

	// Test whether the absence of panic calls is properly detected.
	if nil != CheckPanic(func() {}) {
		t.Error("False positive panic detection!")
	}
}
