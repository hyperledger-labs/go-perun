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

package client

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
)

func TestUpdateResponder_Accept_NilArgs(t *testing.T) {
	err := new(UpdateResponder).Accept(nil) // nolint: staticcheck
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestUpdateResponder_Reject_NilArgs(t *testing.T) {
	err := new(UpdateResponder).Reject(nil, "reason") // nolint: staticcheck
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

// Channel.Update() is defined in `client/update.go` so its test can be found
// here as well.
func TestChannel_Update_NilArgs(t *testing.T) {
	err := new(Channel).Update(nil, new(channel.State)) // nolint: staticcheck
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestRequestTimedOutError(t *testing.T) {
	err := newRequestTimedOutError("", "")
	requestTimedOutError := RequestTimedOutError("")

	t.Run("direct_error", func(t *testing.T) {
		gotRequestTimedOutError := errors.As(err, &requestTimedOutError)
		require.True(t, gotRequestTimedOutError)
	})

	t.Run("wrapped_error", func(t *testing.T) {
		wrappedError := errors.WithMessage(err, "some higher level error")
		gotRequestTimedOutError := errors.As(wrappedError, &requestTimedOutError)
		require.True(t, gotRequestTimedOutError)
	})
}
