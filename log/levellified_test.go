// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package log

import (
	"bytes"
	"log"
	"testing"
)

const (
	logStr = "Perun"
)

func TestLevellified(t *testing.T) {
	buf := new(bytes.Buffer)
	stdLogger := log.New(buf, "", 0)
	ll := Levellified{StdLogger: stdLogger}

	// prefix testing closure
	requirePrefix := func(logFn func(...interface{}), lvl Level) {
		buf.Reset()
		logFn(logStr)
		golden := "[" + lvl.String() + "] " + logStr + "\n"
		if buf.String() != golden {
			t.Errorf("Want: %q, have: %q", golden, buf.String())
		}
	}

	// no logging testing closure
	requireSilent := func(logFn func(...interface{})) {
		buf.Reset()
		logFn(logStr)
		if buf.Len() > 0 {
			t.Errorf("Want no logging, have: %q", buf.String())
		}
	}

	// full test for Print and Println type methods
	testLogger := func(logFn func(...interface{}), lvl Level) {
		// should not log in range [Error..lvl-1]
		for ll.Lvl = ErrorLevel; ll.Lvl < lvl; ll.Lvl++ {
			requireSilent(logFn)
		}
		// should log in range [lvl..Trace]
		for ; ll.Lvl <= TraceLevel; ll.Lvl++ {
			requirePrefix(logFn, lvl)
		}
	}

	// full test for Printf type methods
	testFLogger := func(logFn func(string, ...interface{}), lvl Level) {
		logWrapper := func(args ...interface{}) { logFn("%s", args...) }
		testLogger(logWrapper, lvl)
	}

	testLogger(ll.Trace, TraceLevel)
	testLogger(ll.Traceln, TraceLevel)
	testFLogger(ll.Tracef, TraceLevel)

	testLogger(ll.Debug, DebugLevel)
	testLogger(ll.Debugln, DebugLevel)
	testFLogger(ll.Debugf, DebugLevel)

	testLogger(ll.Info, InfoLevel)
	testLogger(ll.Infoln, InfoLevel)
	testFLogger(ll.Infof, InfoLevel)

	testLogger(ll.Warn, WarnLevel)
	testLogger(ll.Warnln, WarnLevel)
	testFLogger(ll.Warnf, WarnLevel)

	testLogger(ll.Error, ErrorLevel)
	testLogger(ll.Errorln, ErrorLevel)
	testFLogger(ll.Errorf, ErrorLevel)

	// note: Panic and Fatal don't need to be tested as those are just taken from
	// the StdLogger itself.
}
