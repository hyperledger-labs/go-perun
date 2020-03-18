// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package logrus

import (
	"encoding/hex"
	"fmt"

	"github.com/sirupsen/logrus"

	"perun.network/go-perun/log"
)

// Logger wraps a logrus logger.
type Logger struct {
	*logrus.Entry
}

var _ log.Logger = (*Logger)(nil)

// FromLogrus creates a logger from a logrus.Logger.
func FromLogrus(l *logrus.Logger) *Logger {
	return &Logger{logrus.NewEntry(l)}
}

// WithField calls WithField on the logrus.Logger.
func (l *Logger) WithField(key string, value interface{}) log.Logger {
	return &Logger{l.Entry.WithField(key, stringify(value))}
}

func stringify(value interface{}) interface{} {
	if str, ok := value.(fmt.Stringer); ok {
		return str.String()
	}
	switch v := value.(type) {
	case [32]byte:
		return hex.EncodeToString(v[:])
	default:
		return value
	}
}

// WithFields calls WithField for all passed fields.
func (l *Logger) WithFields(fields log.Fields) (ret log.Logger) {
	newFields := make(map[string]interface{})
	for k, v := range fields {
		newFields[k] = stringify(v)
	}
	return &Logger{l.Entry.WithFields(newFields)}
}

// WithError calls WithError on the logrus.Logger.
func (l *Logger) WithError(e error) log.Logger {
	return &Logger{l.Entry.WithError(e)}
}

// Set sets a logrus logger as the current framework logger with the given level
// and formatter.
func Set(level logrus.Level, formatter logrus.Formatter) {
	logger := logrus.New()
	logger.SetLevel(level)
	logger.SetFormatter(formatter)
	log.Set(FromLogrus(logger))
}
