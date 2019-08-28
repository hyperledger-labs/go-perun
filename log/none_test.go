// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNone tests the none logger for coverage :)
func TestNone(t *testing.T) {
	None := &none{}

	None.Printf("")
	None.Print()
	None.Println()
	None.Tracef("")
	None.Debugf("")
	None.Infof("")
	None.Warnf("")
	None.Errorf("")
	None.Trace()
	None.Debug()
	None.Info()
	None.Warn()
	None.Error()
	None.Traceln()
	None.Debugln()
	None.Infoln()
	None.Warnln()
	None.Errorln()

	a := assert.New(t)

	// Monkey patch the exit function
	code := 0
	exit = func(i int) {
		code = i
	}
	// Test fatal functions
	funs := []func(){func() { None.Fatalf("") }, func() { None.Fatal() }, func() { None.Fatalln() }}
	for _, f := range funs {
		code = 0
		f()
		a.Equal(1, code)
	}

	a.Panics(func() { None.Panicf("") })
	a.Panics(func() { None.Panic() })
	a.Panics(func() { None.Panicln() })

	a.Equal(None.WithField("", ""), None)
	a.Equal(None.WithFields(nil), None)
	a.Equal(None.WithError(nil), None)
}
