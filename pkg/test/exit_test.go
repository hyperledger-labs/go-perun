// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"os"
	"testing"
)

func TestCheckExit(t *testing.T) {
	// Test whether an exit call is correctly detected and whether the status
	// code is properly recorded.
	exit, code := CheckExit(func() { os.Exit(1) })
	if !exit {
		t.Error("Failed to detect os.Exit call!")
	}
	if code != 1 {
		t.Errorf("Detected wrong status code: expected %v, but was %v", 1, code)
	}

	// Test whether absent exit calls are correctly detected as well.
	exit, _ = CheckExit(func() {})
	if exit {
		t.Error("False positive os.Exit detection!")
	}
}
