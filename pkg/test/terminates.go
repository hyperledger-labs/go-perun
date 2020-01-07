// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"context"
	"time"
)

// TerminatesCtx checks whether a function terminates before a context is done.
func TerminatesCtx(ctx context.Context, fn func()) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	done := make(chan struct{}, 1)
	go func() {
		fn()
		done <- struct{}{}
	}()

	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}

// Terminates checks whether a function terminates within a certain timeout.
func Terminates(deadline time.Duration, fn func()) bool {
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()
	return TerminatesCtx(ctx, fn)
}

// AssertTerminatesCtx asserts that a function terminates before a context is
// done.
func AssertTerminatesCtx(ctx context.Context, t T, fn func()) {
	t.Helper()

	if !TerminatesCtx(ctx, fn) {
		t.Error("function should have terminated within deadline")
	}
}

// AssertNotTerminatesCtx asserts that a function does not terminate before a
// context is done.
func AssertNotTerminatesCtx(ctx context.Context, t T, fn func()) {
	t.Helper()

	if TerminatesCtx(ctx, fn) {
		t.Error("Function should not have terminated within deadline")
	}
}

// AssertTerminates asserts that a function terminates within a certain
// timeout.
func AssertTerminates(t T, deadline time.Duration, fn func()) {
	t.Helper()

	if !Terminates(deadline, fn) {
		t.Error("Function should have terminated within deadline")
	}
}

// AssertNotTerminates asserts that a function does not terminate within a
// certain timeout.
func AssertNotTerminates(t T, deadline time.Duration, fn func()) {
	t.Helper()

	if Terminates(deadline, fn) {
		t.Error("Function should not have terminated within deadline")
	}
}
