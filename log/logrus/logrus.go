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
