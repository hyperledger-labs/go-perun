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

package client

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/watcher"
	"perun.network/go-perun/wire"
	perunsync "polycry.pt/poly-go/sync"
)

// Channel is the channel controller, progressing the channel state machine and
// executing the channel update and dispute protocols.
//
// Currently, only the two-party protocol is fully implemented.
type Channel struct {
	perunsync.OnCloser
	log.Embedding

	client      *Client
	conn        *channelConn
	machine     persistence.StateMachine
	machMtx     perunsync.Mutex
	statesPub   watcher.StatesPub
	onUpdate    func(from, to *channel.State)
	adjudicator channel.Adjudicator
	wallet      wallet.Wallet

	parent                *Channel            // must be nil for ledger channel
	subChannelFundings    *updateInterceptors // awaited subchannel funding updates
	subChannelWithdrawals *updateInterceptors // awaited subchannel settlement updates
}

// newChannel is internally used by the Client to create a new channel
// controller after the channel proposal protocol ran successfully.
func (c *Client) newChannel(
	acc wallet.Account,
	parent *Channel,
	peers []map[int]wire.Address, // peerIdx, BackendID -> Address
	params channel.Params,
) (*Channel, error) {
	machine, err := channel.NewStateMachine(acc, params)
	if err != nil {
		return nil, errors.WithMessage(err, "creating state machine")
	}
	return c.channelFromMachine(machine, parent, peers)
}

// channelFromSource is used to create a channel controller from restored data.
func (c *Client) channelFromSource(s channel.Source, parent *Channel, peers []map[int]wire.Address) (*Channel, error) {
	accs, err := c.wallet.Unlock(s.Params().Parts[s.Idx()])
	if err != nil {
		return nil, errors.WithMessage(err, "unlocking account for channel")
	}

	machine, err := channel.RestoreStateMachine(accs, s)
	if err != nil {
		return nil, errors.WithMessage(err, "restoring state machine")
	}

	return c.channelFromMachine(machine, parent, peers)
}

// channelFromMachine creates a channel controller around the passed state machine.
func (c *Client) channelFromMachine(machine *channel.StateMachine, parent *Channel, peers []map[int]wire.Address) (*Channel, error) {
	logger := c.logChan(machine.ID())
	machine.SetLog(logger) // client logger has more fields
	pmachine := persistence.FromStateMachine(machine, c.pr)

	// bundle peers into channel connection
	conn, err := newChannelConn(machine.ID(), peers, machine.Idx(), &c.conn, &c.conn)
	if err != nil {
		return nil, errors.WithMessagef(err, "setting up channel connection")
	}

	conn.SetLog(logger)
	return &Channel{
		client:                c,
		parent:                parent,
		statesPub:             noopStatesPub{},
		OnCloser:              conn,
		Embedding:             log.MakeEmbedding(logger),
		conn:                  conn,
		machine:               pmachine,
		adjudicator:           c.adjudicator,
		wallet:                c.wallet,
		subChannelFundings:    newUpdateInterceptors(),
		subChannelWithdrawals: newUpdateInterceptors(),
	}, nil
}

// Close closes the channel and all associated peer subscriptions.
func (c *Channel) Close() error {
	return c.conn.Close()
}

// IsClosed returns whether the channel is closed.
func (c *Channel) IsClosed() bool {
	return c.conn.r.IsClosed()
}

// Ctx returns a context that is active for the channel's lifetime.
func (c *Channel) Ctx() context.Context {
	return c.conn.r.Ctx()
}

func (c *Channel) logPeer(idx channel.Index) log.Logger {
	return c.Log().WithField("peerIdx", idx)
}

// ID returns the channel ID.
func (c *Channel) ID() map[int]channel.ID {
	return c.machine.ID()
}

// Idx returns our index in the channel.
func (c *Channel) Idx() channel.Index {
	return c.machine.Idx()
}

// Params returns the channel parameters.
func (c *Channel) Params() *channel.Params {
	return c.machine.Params()
}

// State returns a pointer to the current state.
// Clone it if you want to modify it.
// Can not be called from an update handler.
func (c *Channel) State() *channel.State {
	c.machMtx.Lock()
	defer c.machMtx.Unlock()

	return c.state()
}

// state returns a pointer to the current state.
// Assumes that the machine mutex has been locked.
// Clone the state if you want to modify it.
func (c *Channel) state() *channel.State {
	return c.machine.State()
}

// Phase returns the current phase of the channel state machine.
// Can not be called from an update handler.
func (c *Channel) Phase() channel.Phase {
	c.machMtx.Lock()
	defer c.machMtx.Unlock()

	return c.machine.Phase()
}

// Peers returns the Perun network addresses of all peers, in the order
// of the channel participants.
func (c *Channel) Peers() []map[int]wire.Address {
	return c.conn.Peers()
}

// Parent returns the parent channel. Can be nil.
func (c *Channel) Parent() *Channel {
	return c.parent
}

// HasApp returns whether the channel has an app.
func (c *Channel) HasApp() bool {
	return !channel.IsNoApp(c.State().App)
}

// init brings the state machine into the InitSigning phase. It is not callable
// by the user since the Client initializes the channel controller.
// The state machine is not locked as this function is expected to be called
// during the initialization phase of the channel controller.
func (c *Channel) init(ctx context.Context, initBals *channel.Allocation, initData channel.Data) error {
	return c.machine.Init(ctx, *initBals, initData)
}

// initExchangeSigsAndEnable exchanges signatures on the initial state.
// The state machine is not locked as this function is expected to be called
// during the initialization phase of the channel controller.
func (c *Channel) initExchangeSigsAndEnable(ctx context.Context) error {
	sig, err := c.machine.Sig(ctx)
	if err != nil {
		return err
	}

	resRecv, err := c.conn.NewUpdateResRecv(0)
	if err != nil {
		return errors.WithMessage(err, "creating update response receiver")
	}
	defer resRecv.Close()

	send := make(chan error)
	go func() {
		send <- c.conn.Send(ctx, &ChannelUpdateAccMsg{
			ChannelID: c.ID(),
			Version:   0,
			Sig:       sig,
		})
	}()

	pidx, cm, err := resRecv.Next(ctx)
	if err != nil {
		return errors.WithMessage(err, "receiving initial state sig")
	}
	acc, ok := cm.(*ChannelUpdateAccMsg)
	if !ok {
		return errors.Errorf(
			"received unexpected message of type (%T) from peer[%d]: %v",
			cm, pidx, cm)
	}

	if err := c.machine.AddSig(ctx, pidx, acc.Sig); err != nil {
		return err
	}
	if err := c.machine.EnableInit(ctx); err != nil {
		return err
	}

	return errors.WithMessage(<-send, "sending initial signature")
}

func (c *Channel) hasLockedFunds() bool {
	return len(c.machine.State().Locked) > 0
}

func sortPeers(peer map[int][]wire.Address) []map[int]wire.Address {
	peerNum := 0
	for i := range peer {
		peerNum = len(peer[i])
	}
	peerSlice := make([]map[int]wire.Address, peerNum)
	for p := 0; p < peerNum; p++ {
		peerSlice[p] = GetPeerMapWire(peer, p)
	}
	return peerSlice
}
