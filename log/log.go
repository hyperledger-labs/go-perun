// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package log implements the logger interface of go-perun. Users are expected
// to pass an implementation of this interface to harmonize go-perun's logging
// with their application logging.
//
// It mimics the interface of logrus, which is go-perun's logger of choice
// It is also possible to pass a simpler logger like the standard library's log
// logger by converting it to a perun logger. Use the Fieldify and Levellify
// factories for that.
package log // import "perun.network/go-perun/log"

import "log"

var (
	// compile-time check that log.Logger implements a StdLogger
	_ StdLogger = &log.Logger{}

	// Log is the framework logger. Framework users should set this variable to
	// their logger. It is set to the None non-logging logger by default.
	Log Logger = None
)

// StdLogger describes the interface of the standard library log package logger.
// It is the base for more complex loggers. A StdLogger can be converted into a
// LevelLogger by wrapping it with a Levellified struct.
type StdLogger interface {
	Printf(format string, args ...interface{})
	Print(...interface{})
	Println(...interface{})

	Fatalf(format string, args ...interface{})
	Fatal(...interface{})
	Fatalln(...interface{})

	Panicf(format string, args ...interface{})
	Panic(...interface{})
	Panicln(...interface{})
}

// LevelLogger is an extension to the StdLogger with different verbosity levels.
type LevelLogger interface {
	StdLogger

	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Trace(...interface{})
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})

	Traceln(...interface{})
	Debugln(...interface{})
	Infoln(...interface{})
	Warnln(...interface{})
	Errorln(...interface{})
}

// Fields is a collection of fields that can be passed to FieldLogger.WithFields
type Fields map[string]interface{}

// Logger is a LevelLogger with structured field logging capabilities.
// This is the interface that needs to be passed to go-perun.
type Logger interface {
	LevelLogger

	WithField(key string, value interface{}) Logger
	WithFields(Fields) Logger
	WithError(error) Logger
}
