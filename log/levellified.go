// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package log

import "log"

// Default level shift, such that Level value 0 is the default log level
const defaultLogLevelShift = 3 // Warn

// compile-time check that Levellified extends a StdLogger to a LevelLogger
var _ LevelLogger = &Levellified{StdLogger: &log.Logger{}}

// Levellified levellifies a standard logger. Calls are just forwarded to the wrapped
// logger's Print{,f,ln} methods with the prefix [level], except for levels
// Panic and Fatal, which are forwarded to the respective methods.
type Levellified struct {
	// wrapped logger
	StdLogger
	// Lvl is the current logging level
	Lvl Level
}

// Level is the log level.
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
	// Warn calls the wrapped logger's Print method with the prefix "[warn]".
	// It is the default level.
	Warn
	// Info calls the wrapped logger's Print method with the prefix "[info]".
	Info
	// Debug calls the wrapped logger's Print method with the prefix "[debug]".
	Debug
	// Trace calls the wrapped logger's Print method with the prefix "[trace]".
	Trace
)

// String returns the string representation.
func (l Level) String() string {
	return [...]string{"fatal", "panic", "error", "warn", "info", "debug", "trace"}[l+defaultLogLevelShift]
}

// Tracef implementents log level trace and format parameters.
func (l *Levellified) Tracef(format string, args ...interface{}) {
	l.lprintf(Trace, format, args...)
}

// Trace implements log level trace.
func (l *Levellified) Trace(args ...interface{}) {
	if l.Lvl >= Trace {
		l.StdLogger.Print(prepend("[trace] ", args)...)
	}
}

// Traceln implements log.TraceLn with white spaces in between arguments.
func (l *Levellified) Traceln(args ...interface{}) {
	if l.Lvl >= Trace {
		l.StdLogger.Println(prepend("[trace]", args)...)
	}
}

// Debugf implementents log level debug and format parameters.
func (l *Levellified) Debugf(format string, args ...interface{}) {
	l.lprintf(Debug, format, args...)
}

// Debug implements log level debug.
func (l *Levellified) Debug(args ...interface{}) {
	if l.Lvl >= Debug {
		l.StdLogger.Print(prepend("[debug] ", args)...)
	}
}

// Debugln implements log.Debugln with white spaces in between arguments.
func (l *Levellified) Debugln(args ...interface{}) {
	if l.Lvl >= Debug {
		l.StdLogger.Println(prepend("[debug]", args)...)
	}
}

// Infof implementents log level info and format parameters.
func (l *Levellified) Infof(format string, args ...interface{}) {
	l.lprintf(Info, format, args...)
}

// Info implements log level info.
func (l *Levellified) Info(args ...interface{}) {
	if l.Lvl >= Info {
		l.StdLogger.Print(prepend("[info] ", args)...)
	}
}

// Infoln implements log.Infoln with white spaces in between arguments.
func (l *Levellified) Infoln(args ...interface{}) {
	if l.Lvl >= Info {
		l.StdLogger.Println(prepend("[info]", args)...)
	}
}

// Warnf implementents log level warn and format parameters.
func (l *Levellified) Warnf(format string, args ...interface{}) {
	l.lprintf(Warn, format, args...)
}

// Warn implements log level warn.
func (l *Levellified) Warn(args ...interface{}) {
	if l.Lvl >= Warn {
		l.StdLogger.Print(prepend("[warn] ", args)...)
	}
}

// Warnln implements log.Warnln with white spaces in between arguments.
func (l *Levellified) Warnln(args ...interface{}) {
	if l.Lvl >= Warn {
		l.StdLogger.Println(prepend("[warn]", args)...)
	}
}

// Errorf implementents log level error and format parameters.
func (l *Levellified) Errorf(format string, args ...interface{}) {
	l.lprintf(Error, format, args...)
}

// Error implements log level error.
func (l *Levellified) Error(args ...interface{}) {
	if l.Lvl >= Error {
		l.StdLogger.Print(prepend("[error] ", args)...)
	}
}

// Errorln implements log.Errorln with white spaces in between arguments.
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
