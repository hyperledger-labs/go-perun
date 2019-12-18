// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package log

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/test"
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

	// Test that fatal functions call exit
	et := test.NewExit(t, &exit)
	funs := []func(){func() { None.Fatalf("") }, func() { None.Fatal() }, func() { None.Fatalln() }}
	for _, fn := range funs {
		et.AssertExit(fn, 1)
	}

	a := assert.New(t)
	a.Panics(func() { None.Panicf("") })
	a.Panics(func() { None.Panic() })
	a.Panics(func() { None.Panicln() })

	a.Same(None.WithField("", ""), None)
	a.Same(None.WithFields(nil), None)
	a.Same(None.WithError(nil), None)
}
