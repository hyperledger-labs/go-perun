// Copyright 2020 - See NOTICE file for copyright holders.
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

// IsDone returns whether ctx is done.
func IsDone(ctx interface{ Done() <-chan struct{} }) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
