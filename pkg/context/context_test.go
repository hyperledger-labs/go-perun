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

package context_test

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	pcontext "perun.network/go-perun/pkg/context"
)

func TestIsContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.True(t, pcontext.IsContextError(errors.WithStack(ctx.Err())))

	// context that immediately times out
	ctx, cancel = context.WithTimeout(context.Background(), 0)
	defer cancel()
	assert.True(t, pcontext.IsContextError(errors.WithStack(ctx.Err())))

	assert.False(t, pcontext.IsContextError(errors.New("no context error")))
}
