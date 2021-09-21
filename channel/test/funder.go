// Copyright 2021 - See NOTICE file for copyright holders.
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

package test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"perun.network/go-perun/channel"
)

// Funder represents a funder setup used for testing.
//
// Funders should return a set of funders used for testing.
// NewFundingRequests should return a new set of funding requests for the given funders.
type Funder interface {
	Funders() []channel.Funder
	NewFundingRequests(context.Context, *testing.T, *rand.Rand) []channel.FundingReq
}

// TestFunder runs a set of generic funder tests.
func TestFunder(ctx context.Context, t *testing.T, rng *rand.Rand, f Funder) {
	funders := f.Funders()
	requests := f.NewFundingRequests(ctx, t, rng)

	errs := make(chan error, len(funders))
	for i := range funders {
		go func(funder channel.Funder, req channel.FundingReq) {
			errs <- funder.Fund(ctx, req)
		}(funders[i], requests[i])
	}

	for range funders {
		select {
		case err := <-errs:
			require.NoError(t, err, "funding should work")
		case <-ctx.Done():
			t.Error(ctx.Err())
		}
	}
}
