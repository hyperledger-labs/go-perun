// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package logrus

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestLogrus(t *testing.T) {
	a := assert.New(t)
	logger, hook := test.NewNullLogger()
	FromLogrus(logger).Println("Anton Ausdemhaus")

	a.Equal(len(hook.Entries), 1)
	a.Equal(hook.LastEntry().Level, logrus.InfoLevel)
	a.Equal(hook.LastEntry().Message, "Anton Ausdemhaus")

	// test WithField
	logger, hook = test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)
	FromLogrus(logger).WithField("field", 123456).Debugln("Bertha Bremsweg")
	a.Equal(len(hook.Entries), 1)
	a.Equal(hook.LastEntry().Level, logrus.DebugLevel)
	a.Equal(hook.LastEntry().Message, "Bertha Bremsweg")
	a.Contains(hook.LastEntry().Data, "field")
	a.Equal(hook.LastEntry().Data["field"], 123456)

	// test WithFields
	logger, hook = test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)
	fields := map[string]interface{}{
		"mars":    249,
		"jupiter": 816,
		"saturn":  1514,
	}
	FromLogrus(logger).WithFields(fields).Errorln("Christian Chaos")
	a.Equal(len(hook.Entries), 1)
	a.Equal(hook.LastEntry().Level, logrus.ErrorLevel)
	a.Equal(hook.LastEntry().Message, "Christian Chaos")
	a.EqualValues(hook.LastEntry().Data, fields)

	// test WithError
	e := errors.New("error-message")
	buf := new(bytes.Buffer)
	FromLogrus(&logrus.Logger{
		Out:       buf,
		Formatter: new(logrus.TextFormatter),
		Hooks:     nil,
		Level:     logrus.DebugLevel,
	}).WithError(e).Warnln("Doris Day")
	a.Contains(buf.String(), "Doris Day")
	a.Contains(buf.String(), "error-message")
}
