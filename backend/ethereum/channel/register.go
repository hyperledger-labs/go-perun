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

package channel

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
)

// Register registers a state on-chain.
// If the state is a final state, register becomes a no-op.
func (a *Adjudicator) Register(ctx context.Context, req channel.AdjudicatorReq, subChannels []channel.SignedState) error {
	if req.Tx.State.IsFinal {
		return a.registerFinal(ctx, req)
	}
	return a.registerNonFinal(ctx, req, subChannels)
}

// registerFinal registers a final state. It ensures that the final state is
// concluded on the adjudicator conctract.
func (a *Adjudicator) registerFinal(ctx context.Context, req channel.AdjudicatorReq) error {
	// In the case of final states, we already call concludeFinal on the
	// adjudicator. Method ensureConcluded calls concludeFinal for final states.
	if err := a.ensureConcluded(ctx, req, nil); err != nil {
		return errors.WithMessage(err, "ensuring Concluded")
	}

	return nil
}

func (a *Adjudicator) registerNonFinal(ctx context.Context, req channel.AdjudicatorReq, subChannels []channel.SignedState) error {
	if err := a.callRegister(ctx, req, subChannels); err != nil {
		return errors.WithMessage(err, "calling register")
	}
	return nil
}
