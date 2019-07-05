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
		for ll.Lvl = Error; ll.Lvl < lvl; ll.Lvl++ {
			requireSilent(logFn)
		}
		// should log in range [lvl..Trace]
		for ; ll.Lvl <= Trace; ll.Lvl++ {
			requirePrefix(logFn, lvl)
		}
	}

	// full test for Printf type methods
	testFLogger := func(logFn func(string, ...interface{}), lvl Level) {
		logWrapper := func(args ...interface{}) { logFn("%s", args...) }
		testLogger(logWrapper, lvl)
	}

	testLogger(ll.Trace, Trace)
	testLogger(ll.Traceln, Trace)
	testFLogger(ll.Tracef, Trace)

	testLogger(ll.Debug, Debug)
	testLogger(ll.Debugln, Debug)
	testFLogger(ll.Debugf, Debug)

	testLogger(ll.Info, Info)
	testLogger(ll.Infoln, Info)
	testFLogger(ll.Infof, Info)

	testLogger(ll.Warn, Warn)
	testLogger(ll.Warnln, Warn)
	testFLogger(ll.Warnf, Warn)

	testLogger(ll.Error, Error)
	testLogger(ll.Errorln, Error)
	testFLogger(ll.Errorf, Error)

	// note: Panic and Fatal don't need to be tested as those are just taken from
	// the StdLogger itself.
}
