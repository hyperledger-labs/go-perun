// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package log

import "log"

// Default level shift, such that Level value 0 is the default log level
const defaultLogLevelShift = 3 // Warn

// compile-time check that Levellified extends a StdLogger to a LevelLogger
var _ LevelLogger = &Levellified{StdLogger: &log.Logger{}}

// Levellifies a standard logger. Calls are just forwarded to the wrapped
// logger's Print{,f,ln} methods with the prefix [level], except for levels
// Panic and Fatal, which are forwarded to the respective methods.
type Levellified struct {
	// wrapped logger
	StdLogger
	// Lvl is the current logging level
	Lvl Level
}

type Level int8

const (
	// Fatal calls the wrapped logger's Fatal method with the prefix "[fatal]".
	// The wrapped logger should usually immediately exit the program with
	// os.Exit(1) or similar.  It it the highest level of severity.
	Fatal Level = iota - defaultLogLevelShift // -3: default level Warn
	// Panic calls the wrapped logger's Panic method with the prefix "[panic]".
	// The wrapped logger should usually call panic with the given message
	Panic
	// Error calls the wrapped logger's Print method with the prefix "[error]".
	Error
	// warn calls the wrapped logger's Print method with the prefix "[warn]".
	// It is the default level.
	Warn
	// info calls the wrapped logger's Print method with the prefix "[info]".
	Info
	// debug calls the wrapped logger's Print method with the prefix "[debug]".
	Debug
	// trace calls the wrapped logger's Print method with the prefix "[trace]".
	Trace
)

func (l Level) String() string {
	return [...]string{"fatal", "panic", "error", "warn", "info", "debug", "trace"}[l+defaultLogLevelShift]
}

func (l *Levellified) Tracef(format string, args ...interface{}) {
	l.lprintf(Trace, format, args...)
}

func (l *Levellified) Trace(args ...interface{}) {
	if l.Lvl >= Trace {
		l.StdLogger.Print(prepend("[trace] ", args)...)
	}
}

func (l *Levellified) Traceln(args ...interface{}) {
	if l.Lvl >= Trace {
		l.StdLogger.Println(prepend("[trace]", args)...)
	}
}

func (l *Levellified) Debugf(format string, args ...interface{}) {
	l.lprintf(Debug, format, args...)
}

func (l *Levellified) Debug(args ...interface{}) {
	if l.Lvl >= Debug {
		l.StdLogger.Print(prepend("[debug] ", args)...)
	}
}

func (l *Levellified) Debugln(args ...interface{}) {
	if l.Lvl >= Debug {
		l.StdLogger.Println(prepend("[debug]", args)...)
	}
}

func (l *Levellified) Infof(format string, args ...interface{}) {
	l.lprintf(Info, format, args...)
}

func (l *Levellified) Info(args ...interface{}) {
	if l.Lvl >= Info {
		l.StdLogger.Print(prepend("[info] ", args)...)
	}
}

func (l *Levellified) Infoln(args ...interface{}) {
	if l.Lvl >= Info {
		l.StdLogger.Println(prepend("[info]", args)...)
	}
}

func (l *Levellified) Warnf(format string, args ...interface{}) {
	l.lprintf(Warn, format, args...)
}

func (l *Levellified) Warn(args ...interface{}) {
	if l.Lvl >= Warn {
		l.StdLogger.Print(prepend("[warn] ", args)...)
	}
}

func (l *Levellified) Warnln(args ...interface{}) {
	if l.Lvl >= Warn {
		l.StdLogger.Println(prepend("[warn]", args)...)
	}
}

func (l *Levellified) Errorf(format string, args ...interface{}) {
	l.lprintf(Error, format, args...)
}

func (l *Levellified) Error(args ...interface{}) {
	if l.Lvl >= Error {
		l.StdLogger.Print(prepend("[error] ", args)...)
	}
}

func (l *Levellified) Errorln(args ...interface{}) {
	if l.Lvl >= Error {
		l.StdLogger.Println(prepend("[error]", args)...)
	}
}

func (l *Levellified) lprintf(lvl Level, format string, args ...interface{}) {
	if l.Lvl >= lvl {
		l.StdLogger.Printf("[%v] "+format, prepend(lvl, args)...)
	}
}

// prepend prepends a slice args with element pre
func prepend(pre interface{}, args []interface{}) []interface{} {
	return append([]interface{}{pre}, args...)
}
