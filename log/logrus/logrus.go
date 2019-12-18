// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package logrus

import (
	log "github.com/sirupsen/logrus"

	perunlog "perun.network/go-perun/log"
)

type logrus struct {
	*log.Entry
}

var _ perunlog.Logger = (*logrus)(nil)

func FromLogrus(l *log.Logger) *logrus {
	return &logrus{log.NewEntry(l)}
}

func (l *logrus) WithField(key string, value interface{}) perunlog.Logger {
	return &logrus{l.Entry.WithField(key, value)}
}

func (l *logrus) WithFields(fields perunlog.Fields) perunlog.Logger {
	var fs map[string]interface{}
	fs = fields
	return &logrus{l.Entry.WithFields(fs)}
}

func (l *logrus) WithError(e error) perunlog.Logger {
	return &logrus{l.Entry.WithError(e)}
}
