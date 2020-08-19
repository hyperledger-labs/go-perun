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

	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/test"
)

func TestProposalOpts_isNonce(t *testing.T) {
	// Nil, empty, and app proposal options do not have nonces.
	require.False(t, WithApp(nil, nil).isNonce())
	require.False(t, ProposalOpts{}.isNonce())
	require.False(t, (*ProposalOpts)(nil).isNonce())

	require.True(t, WithNonce(NonceShare{}).isNonce())
	require.True(t, WithNonceFrom(test.Prng(t)).isNonce())
	require.True(t, WithRandomNonce().isNonce())
}
