// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package logrus

import (
	"github.com/sirupsen/logrus"

	"perun.network/go-perun/log"
)

type logger struct {
	*logrus.Entry
}

var _ log.Logger = (*logger)(nil)

func FromLogrus(l *logrus.Logger) *logger {
	return &logger{logrus.NewEntry(l)}
}

func (l *logger) WithField(key string, value interface{}) log.Logger {
	return &logger{l.Entry.WithField(key, value)}
}

func (l *logger) WithFields(fields log.Fields) log.Logger {
	var fs map[string]interface{}
	fs = fields
	return &logger{l.Entry.WithFields(fs)}
}

func (l *logger) WithError(e error) log.Logger {
	return &logger{l.Entry.WithError(e)}
}
}
