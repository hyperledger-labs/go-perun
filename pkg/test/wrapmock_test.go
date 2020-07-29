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

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// spying mock to check if Error function was called.
// If wasCalled is true, it means that an error.
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
// object in the right situations.
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

// TestNewWrapMock tests that NewWrapMock() successfully creates a working WrapMock.
func TestNewWrapMock(t *testing.T) {
	globalFruit = &fruit{
		NewWrapMock(t),
	}
	banana()
	globalFruit.AssertCalled()
}
