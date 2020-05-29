// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

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
