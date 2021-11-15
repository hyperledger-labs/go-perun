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

package log

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"polycry.pt/poly-go/test"
)

// TestNone tests the none logger for coverage :).
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
