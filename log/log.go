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
	Log = None
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

// Printf calls Printf on the global Logger object
func Printf(format string, args ...interface{}) { Log.Printf(format, args) }

// Print calls Print on the global Logger object
func Print(args ...interface{}) { Log.Print(args) }

// Println calls Println on the global Logger object
func Println(args ...interface{}) { Log.Println(args) }

// Fatalf calls Fatalf on the global Logger object
func Fatalf(format string, args ...interface{}) { Log.Fatalf(format, args) }

// Fatal calls Fatal on the global Logger object
func Fatal(args ...interface{}) { Log.Fatal(args) }

// Fatalln calls Fatalln on the global Logger object
func Fatalln(args ...interface{}) { Log.Fatalln(args) }

// Panicf calls Panicf on the global Logger object
func Panicf(format string, args ...interface{}) { Log.Panicf(format, args) }

// Panic calls Panic on the global Logger object
func Panic(args ...interface{}) { Log.Panic(args) }

// Panicln calls Panicln on the global Logger object
func Panicln(args ...interface{}) { Log.Panicln(args) }

// Tracef calls Tracef on the global Logger object
func Tracef(format string, args ...interface{}) { Log.Tracef(format, args) }

// Trace calls Trace on the global Logger object
func Trace(args ...interface{}) { Log.Trace(args) }

// Traceln calls Traceln on the global Logger object
func Traceln(args ...interface{}) { Log.Traceln(args) }

// Debugf calls Debugf on the global Logger object
func Debugf(format string, args ...interface{}) { Log.Debugf(format, args) }

// Debug calls Debug on the global Logger object
func Debug(args ...interface{}) { Log.Debug(args) }

// Debugln calls Debugln on the global Logger object
func Debugln(args ...interface{}) { Log.Debugln(args) }

// Infof calls Infof on the global Logger object
func Infof(format string, args ...interface{}) { Log.Infof(format, args) }

// Info calls Info on the global Logger object
func Info(args ...interface{}) { Log.Info(args) }

// Infoln calls Infoln on the global Logger object
func Infoln(args ...interface{}) { Log.Infoln(args) }

// Warnf calls Warnf on the global Logger object
func Warnf(format string, args ...interface{}) { Log.Warnf(format, args) }

// Warn calls Warn on the global Logger object
func Warn(args ...interface{}) { Log.Warn(args) }

// Warnln calls Warnln on the global Logger object
func Warnln(args ...interface{}) { Log.Warnln(args) }

// Errorf calls Errorf on the global Logger object
func Errorf(format string, args ...interface{}) { Log.Errorf(format, args) }

// Error calls Error on the global Logger object
func Error(args ...interface{}) { Log.Error(args) }

// Errorln calls Errorln on the global Logger object
func Errorln(args ...interface{}) { Log.Errorln(args) }
