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
	"bytes"
	"encoding/hex"
	"errors"
	"math/rand"
	"testing"

	_ "perun.network/go-perun/backend/sim" // backend init
	"perun.network/go-perun/log"
	wtest "perun.network/go-perun/wallet/test"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestLogrus(t *testing.T) {
	t.Run("Info", testLogrusInfo)
	t.Run("Stringer", testLogrusStringer)
	t.Run("WithError", testLogrusWithError)
	t.Run("WithField", testLogrusWithField)
	t.Run("WithFields", testLogrusWithFields)
}

func testLogrusInfo(t *testing.T) {
	logger, hook := test.NewNullLogger()
	FromLogrus(logger).Println("Anton Ausdemhaus")

	assert.Equal(t, len(hook.Entries), 1)
	assert.Equal(t, hook.LastEntry().Level, logrus.InfoLevel)
	assert.Equal(t, hook.LastEntry().Message, "Anton Ausdemhaus")
}

func testLogrusStringer(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDDDD))
	addr := wtest.NewRandomAddress(rng)
	var data [32]byte
	rng.Read(data[:])
	logger, hook := test.NewNullLogger()
	FromLogrus(logger).WithFields(log.Fields{"addr": addr, "data": data}).Infoln("")

	assert.Contains(t, hook.LastEntry().Data, "addr")
	assert.Equal(t, hook.LastEntry().Data["addr"], addr.String())
	assert.Contains(t, hook.LastEntry().Data, "data")
	assert.Equal(t, hook.LastEntry().Data["data"], hex.EncodeToString(data[:]))
}

func testLogrusWithError(t *testing.T) {
	e := errors.New("error-message")
	buf := new(bytes.Buffer)
	FromLogrus(&logrus.Logger{
		Out:       buf,
		Formatter: new(logrus.TextFormatter),
		Hooks:     nil,
		Level:     logrus.DebugLevel,
	}).WithError(e).Warnln("Doris Day")

	assert.Contains(t, buf.String(), "Doris Day")
	assert.Contains(t, buf.String(), "error-message")
}

func testLogrusWithField(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)
	FromLogrus(logger).WithField("field", 123456).Debugln("Bertha Bremsweg")

	assert.Equal(t, len(hook.Entries), 1)
	assert.Equal(t, hook.LastEntry().Level, logrus.DebugLevel)
	assert.Equal(t, hook.LastEntry().Message, "Bertha Bremsweg")
	assert.Contains(t, hook.LastEntry().Data, "field")
	assert.Equal(t, hook.LastEntry().Data["field"], 123456)
}

func testLogrusWithFields(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)
	fields := map[string]interface{}{
		"mars":    249,
		"jupiter": 816,
		"saturn":  1514,
	}
	FromLogrus(logger).WithFields(fields).Errorln("Christian Chaos")

	assert.Equal(t, len(hook.Entries), 1)
	assert.Equal(t, hook.LastEntry().Level, logrus.ErrorLevel)
	assert.Equal(t, hook.LastEntry().Message, "Christian Chaos")
	assert.EqualValues(t, hook.LastEntry().Data, fields)
}
