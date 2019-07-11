// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"github.com/pkg/errors"
)

type TransitionError struct {
	error
	ID ID
}

func newTransitionError(id ID, msg string) *TransitionError {
	return &TransitionError{
		error: errors.New(msg),
		ID:    id,
	}
}

func IsTransitionError(err error) bool {
	_, ok := err.(*TransitionError)
	return ok
}
