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

package errors_test

import (
	stderrors "errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/context/test"
	"perun.network/go-perun/pkg/errors"
)

func TestGatherer_Failed(t *testing.T) {
	g := errors.NewGatherer()

	select {
	case <-g.Failed():
		t.Fatal("Failed must not be closed")
	default:
	}

	g.Add(stderrors.New(""))

	select {
	case <-g.Failed():
	default:
		t.Fatal("Failed must be closed")
	}
}

func TestGatherer_Go_and_Wait(t *testing.T) {
	g := errors.NewGatherer()

	const timeout = 10 * time.Millisecond

	g.Go(func() error {
		time.Sleep(timeout)
		return stderrors.New("")
	})

	test.AssertNotTerminates(t, timeout/2, g.Wait)
	test.AssertTerminates(t, timeout, g.Wait)
	require.Error(t, g.Err())
}

func TestGatherer_Add_and_Err(t *testing.T) {
	g := errors.NewGatherer()

	require.NoError(t, g.Err())

	g.Add(stderrors.New("1"))
	g.Add(stderrors.New("2"))
	require.Error(t, g.Err())
	require.Len(t, errors.Causes(g.Err()), 2)
}

func TestCauses(t *testing.T) {
	g := errors.NewGatherer()
	require.Len(t, errors.Causes(g.Err()), 0)

	g.Add(stderrors.New("1"))
	require.Len(t, errors.Causes(g.Err()), 1)

	g.Add(stderrors.New("2"))
	require.Len(t, errors.Causes(g.Err()), 2)

	g.Add(stderrors.New("3"))
	require.Len(t, errors.Causes(g.Err()), 3)
}

func TestAccumulatedError_Error(t *testing.T) {
	g := errors.NewGatherer()
	g.Add(stderrors.New("1"))
	require.Equal(t, g.Err().Error(), "(1 error)\n1): 1")

	g.Add(stderrors.New("2"))
	require.Equal(t, g.Err().Error(), "(2 errors)\n1): 1\n2): 2")
}
