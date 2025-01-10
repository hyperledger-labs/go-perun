// Copyright 2022 - See NOTICE file for copyright holders.
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

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wallet"
)

// AlwaysAcceptChannelHandler returns a channel proposal handler that accepts
// all channel proposals.
func AlwaysAcceptChannelHandler(ctx context.Context, addr map[wallet.BackendID]wallet.Address, channels chan *client.Channel, errs chan<- error) client.ProposalHandlerFunc {
	return func(cp client.ChannelProposal, pr *client.ProposalResponder) {
		switch cp := cp.(type) {
		case *client.LedgerChannelProposalMsg:
			ch, err := pr.Accept(ctx, cp.Accept(addr, client.WithRandomNonce()))
			if err != nil {
				errs <- err
			}
			if ch != nil {
				channels <- ch
			}
		default:
			errs <- errors.Errorf("invalid channel proposal: %v", cp)
		}
	}
}

// AlwaysRejectChannelHandler returns a channel proposal handler that rejects
// all channel proposals.
func AlwaysRejectChannelHandler(ctx context.Context, errs chan<- error) client.ProposalHandlerFunc {
	return func(cp client.ChannelProposal, pr *client.ProposalResponder) {
		err := pr.Reject(ctx, "not accepting channels")
		if err != nil {
			errs <- err
		}
	}
}

// AlwaysAcceptUpdateHandler returns a channel update handler that accepts
// all channel updates.
func AlwaysAcceptUpdateHandler(ctx context.Context, errs chan error) client.UpdateHandlerFunc {
	return func(
		s *channel.State, cu client.ChannelUpdate, ur *client.UpdateResponder,
	) {
		err := ur.Accept(ctx)
		if err != nil {
			errs <- errors.WithMessage(err, "accepting channel update")
		}
	}
}

// AlwaysRejectUpdateHandler returns a channel update handler that rejects all
// channel updates.
func AlwaysRejectUpdateHandler(ctx context.Context, errs chan error) client.UpdateHandlerFunc {
	return func(state *channel.State, update client.ChannelUpdate, responder *client.UpdateResponder) {
		err := responder.Reject(ctx, "")
		if err != nil {
			errs <- errors.WithMessage(err, "rejecting channel update")
		}
	}
}
