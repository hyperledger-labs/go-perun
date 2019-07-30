// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// spying mock to check if Error function was called.
// If wasCalled is true, it means that an error
type errorSpy struct {
	t         *testing.T
	wasCalled bool
}

func (s *errorSpy) Error(...interface{}) {
	s.wasCalled = true
}

func (s *errorSpy) Errorf(string, ...interface{}) {
	s.wasCalled = true
}

// called checks whether the spy was called. The called state is reset after a
// call to called().
func (s *errorSpy) called() bool {
	c := s.wasCalled
	s.wasCalled = false
	return c
}

func (s *errorSpy) assertError(msg string) {
	assert.True(s.t, s.called(), msg)
}

func (s *errorSpy) assertNoError(msg string) {
	assert.False(s.t, s.called(), msg)
}

type fruit struct {
	WrapMock
}

func (g *fruit) banana() {
	g.AssertWrapped()
}

var globalFruit *fruit

func banana() {
	globalFruit.banana()
}

func apple() {
	globalFruit.banana()
}

func onion() {}

// TestWrapMock tests that the WrapMock does indeed call Error on the testing
// object in the right situations
func TestWrapMock(t *testing.T) {
	spy := &errorSpy{t: t}
	globalFruit = &fruit{
		WrapMock{t: spy},
	}

	banana()
	spy.assertNoError("global banana() wraps method banana(), no Error should be produced")

	globalFruit.AssertCalled()
	spy.assertNoError("banana() calls method on global, no Error should be produced")

	apple()
	spy.assertError("global apple() doesn't wrap method banana(), Error should be produced")

	globalFruit.AssertCalled()
	spy.assertNoError("apple() calls method on global, no Error should be produced")

	onion()
	globalFruit.AssertCalled()
	spy.assertError("onion() doesn't call method on global, Error should be produced")
}

// TestNewWrapMock tests that NewWrapMock() successfully creates a working WrapMock
func TestNewWrapMock(t *testing.T) {
	globalFruit = &fruit{
		NewWrapMock(t),
	}
	banana()
	globalFruit.AssertCalled()
}
