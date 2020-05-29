// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package msg

import (
	"io"
	"testing"

	"perun.network/go-perun/pkg/io/test"
)

type serializerMsg struct {
	Msg Msg
}

func (msg *serializerMsg) Encode(writer io.Writer) error {
	return Encode(msg.Msg, writer)
}

func (msg *serializerMsg) Decode(reader io.Reader) (err error) {
	msg.Msg, err = Decode(reader)
	return err
}

// TestMsg performs generic tests on a wire.Msg object
func TestMsg(t *testing.T, msg Msg) {
	test.GenericSerializerTest(t, &serializerMsg{msg})
}
