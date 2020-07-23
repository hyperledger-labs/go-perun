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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateResponder_Accept_NilArgs(t *testing.T) {
	err := new(UpdateResponder).Accept(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestUpdateResponder_Reject_NilArgs(t *testing.T) {
	err := new(UpdateResponder).Reject(nil, "reason")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

// Channel.Update() is defined in `client/update.go` so its test can be found
// here as well
func TestChannel_Update_NilArgs(t *testing.T) {
	err := new(Channel).Update(nil, *new(ChannelUpdate))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}
