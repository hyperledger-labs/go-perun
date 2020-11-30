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
func (a *Adjudicator) Register(ctx context.Context, req channel.AdjudicatorReq) (*channel.RegisteredEvent, error) {
	if req.Tx.State.IsFinal {
		return a.registerFinal(ctx, req)
	}
	return a.registerNonFinal(ctx, req)
}

// registerFinal registers a final state. It ensures that the final state is
// concluded on the adjudicator conctract.
func (a *Adjudicator) registerFinal(ctx context.Context, req channel.AdjudicatorReq) (*channel.RegisteredEvent, error) {
	// In the case of final states, we already call concludeFinal on the
	// adjudicator. Method ensureConcluded calls concludeFinal for final states.
	if err := a.ensureConcluded(ctx, req, nil); err != nil {
		return nil, errors.WithMessage(err, "ensuring Concluded")
	}

	return channel.NewRegisteredEvent(
		req.Params.ID(),
		new(channel.ElapsedTimeout), // concludeFinal skips registration
		req.Tx.Version,
	), nil
}

func (a *Adjudicator) registerNonFinal(ctx context.Context, req channel.AdjudicatorReq) (*channel.RegisteredEvent, error) {
	_sub, err := a.Subscribe(ctx, req.Params)
	if err != nil {
		return nil, err
	}
	sub := _sub.(*RegisteredSub)
	// nolint:errcheck
	defer sub.Close()

	// call register if there was no past event
	if !sub.hasPast() {
		if err := a.callRegister(ctx, req); IsErrTxFailed(err) {
			a.log.Warn("Calling register failed, waiting for event anyways...")
		} else if err != nil {
			return nil, errors.WithMessage(err, "calling register")
		}
	}

	// iterate over state registrations and call refute until correct version got
	// registered.
	for {
		switch ev := sub.Next().(type) {
		case nil:
			// the subscription error might be nil, so to ensure a non-nil error, we
			// create a new one.
			return nil, errors.Errorf("subscription closed with error %v", sub.Err())

		case *channel.RegisteredEvent:
			if req.Tx.Version > ev.Version() {
				if err := a.callRefute(ctx, req); IsErrTxFailed(err) {
					a.log.Warn("Calling refute failed, waiting for event anyways...")
				} else if err != nil {
					return nil, errors.WithMessage(err, "calling refute")
				}
				continue // wait for next event
			}
			return ev, nil // version matches, we're done

		case *channel.ProgressedEvent:
			return nil, errors.New("refutation phase already finished")

		case *channel.ConcludedEvent:
			if req.Tx.Version > ev.Version() {
				if err := a.callRefute(ctx, req); IsErrTxFailed(err) {
					a.log.Warn("Calling refute failed, waiting for event anyways...")
				} else if err != nil {
					return nil, errors.WithMessage(err, "calling refute")
				}
				continue // wait for next event
			}
			return channel.NewRegisteredEvent(ev.ID(), ev.Timeout(), ev.Version()), nil // version matches, we're done
		}
	}
}
