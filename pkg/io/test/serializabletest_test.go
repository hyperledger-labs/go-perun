// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"fmt"
	"io"
	"testing"
)

type IncompleteRead struct {
	xs [2]byte
}

func (d *IncompleteRead) Encode(w io.Writer) error {
	_, err := w.Write(d.xs[:])
	return err
}

func (d *IncompleteRead) Decode(r io.Reader) error {
	n, err := r.Read(make([]byte, 1))
	if n != 1 {
		return fmt.Errorf("Expected reading %d bytes, read %d", 1, n)
	}
	return err
}

func MaybeTestGenericDecodeEncodeTest_termination_ShouldFail(t *testing.T) {
	d := IncompleteRead{[2]byte{1, 2}}
	genericDecodeEncodeTest(t, &d)
}
