// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package log

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/test"
)

// none2 is a wrapper around none to have a different Logger type from none for
// testing.
type none2 struct{ none }

func TestSetGet(t *testing.T) {
	l := new(none2)
	Set(l)
	assert.Same(t, l, Get(), "Set(l) should set global logger to l")

	Set(nil)
	assert.IsType(t, &none{}, logger, "Set(nil) should set global logger to none-logger")
}

type wrappedMock struct {
	test.WrapMock
}

// Logger interface

func (m *wrappedMock) Printf(string, ...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Print(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Println(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Fatalf(string, ...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Fatal(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Fatalln(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Panicf(string, ...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Panic(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Panicln(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Tracef(string, ...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Debugf(string, ...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Infof(string, ...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Warnf(string, ...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Errorf(string, ...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Trace(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Debug(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Info(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Warn(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Error(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Traceln(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Debugln(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Infoln(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Warnln(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) Errorln(...interface{}) {
	m.AssertWrapped()
}

func (m *wrappedMock) WithField(string, interface{}) Logger {
	m.AssertWrapped()
	return m
}

func (m *wrappedMock) WithFields(Fields) Logger {
	m.AssertWrapped()
	return m
}

func (m *wrappedMock) WithError(error) Logger {
	m.AssertWrapped()
	return m
}

// compile-time check that wrappedMock implements a Logger
var _ Logger = (*wrappedMock)(nil)

func TestGlobalCalls(t *testing.T) {
	m := &wrappedMock{test.NewWrapMock(t)}
	Set(m)

	Printf("")
	m.AssertCalled()
	Fatalf("")
	m.AssertCalled()
	Panicf("")
	m.AssertCalled()
	Tracef("")
	m.AssertCalled()
	Debugf("")
	m.AssertCalled()
	Infof("")
	m.AssertCalled()
	Warnf("")
	m.AssertCalled()
	Errorf("")
	m.AssertCalled()

	Print()
	m.AssertCalled()
	Fatal()
	m.AssertCalled()
	Panic()
	m.AssertCalled()
	Trace()
	m.AssertCalled()
	Debug()
	m.AssertCalled()
	Info()
	m.AssertCalled()
	Warn()
	m.AssertCalled()
	Error()
	m.AssertCalled()

	Println()
	m.AssertCalled()
	Fatalln()
	m.AssertCalled()
	Panicln()
	m.AssertCalled()
	Traceln()
	m.AssertCalled()
	Debugln()
	m.AssertCalled()
	Infoln()
	m.AssertCalled()
	Warnln()
	m.AssertCalled()
	Errorln()
	m.AssertCalled()

	WithField("", "")
	m.AssertCalled()
	WithFields(nil)
	m.AssertCalled()
	WithError(nil)
	m.AssertCalled()
}
