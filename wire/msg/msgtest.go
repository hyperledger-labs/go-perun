// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import (
	"io"
	"testing"

	"perun.network/go-perun/pkg/io/test"
)

type serializableMsg struct {
	Msg Msg
}

func (msg *serializableMsg) Encode(writer io.Writer) error {
	return Encode(msg.Msg, writer)
}

func (msg *serializableMsg) Decode(reader io.Reader) error {
	var err error
	msg.Msg, err = Decode(reader)
	return err
}

// TestMsg performs generic tests on a wire.Msg object
func TestMsg(t *testing.T, msg Msg) {
	test.GenericSerializableTest(t, &serializableMsg{msg})
}
