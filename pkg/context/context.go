// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

// Package context contains helper utilities regarding go contexts.
package context // import "perun.network/go-perun/pkg/context"

import (
	"context"

	"github.com/pkg/errors"
)

// IsContextError returns whether the given error originates from a context that
// was cancelled or whose deadline exceeded. Prior to checking, the error is
// unwrapped by calling errors.Cause.
func IsContextError(err error) bool {
	err = errors.Cause(err)
	return err == context.Canceled || err == context.DeadlineExceeded
}
