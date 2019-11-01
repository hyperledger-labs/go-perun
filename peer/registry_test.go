// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/backend/sim/wallet"
)

var _ Dialer = (*mockDialer)(nil)

type mockDialer struct {
	dial  chan Conn
	mutex sync.Mutex
}

func (d *mockDialer) Close() error {
	close(d.dial)
	return nil
}

func (d *mockDialer) Dial(ctx context.Context, addr Address) (Conn, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	select {
	case <-ctx.Done():
		return nil, errors.New("aborted manually")
	case conn := <-d.dial:
		if conn != nil {
			return conn, nil
		} else {
			return nil, errors.New("dialer closed")
		}
	}
}

// TestRegistry_Get tests that when calling Get(), existing peers are returned,
// and when unknown peers are requested, a temporary peer is create that is
// dialed in the background. It also tests that the dialing process combines
// with the Listener, so that if a connection to a peer that is still being
// dialed comes in, the peer is assigned that connection.
func TestRegistry_Get(t *testing.T) {
	t.Parallel()
	for i := 0; i < 2; i++ {
		i := i
		t.Run(fmt.Sprintf("subtest %d", i), func(t *testing.T) {
			t.Parallel()
			rng := rand.New(rand.NewSource(0xb0baFEDD))
			d := &mockDialer{dial: make(chan Conn)}
			r := NewRegistry(func(*Peer) {}, d)

			addr := wallet.NewRandomAddress(rng)
			p := r.Get(addr)
			assert.NotNil(t, p, "Get() must not return nil", i)
			assert.Equal(t, p, r.Get(addr), "Get must return the existing peer", i)
			<-time.NewTimer(timeout).C
			assert.NotEqual(t, p, r.Get(wallet.NewRandomAddress(rng)),
				"Get() must return different peers for different addresses", i)

			select {
			case <-p.exists:
				t.Fatal("Peer that is still being dialed must not exist", i)
			default:
			}

			if i == 0 {
				// On the first run, dialing completes normally.
				d.dial <- newMockConn(nil)
			} else {
				// On the second run, a connection comes in before dialing
				// completes.
				r.Register(addr, newMockConn(nil))
			}

			<-time.NewTimer(timeout).C

			select {
			case <-p.exists:
			default:
				t.Fatal("Peer that is successfully dialed must exist", i)
			}

			assert.False(t, p.isClosed(), "Dialed peer must not be closed", i)
		})
	}
}

// TestRegistry_addPeer tests that addPeer() calls the Registry's subscription
// function. Other aspects of the function are already tested in other tests.
func TestRegistry_addPeer_Subscribe(t *testing.T) {
	called := false
	r := NewRegistry(func(*Peer) { called = true }, nil)

	assert.False(t, called, "subscription must not have been called yet")
	r.addPeer(nil, nil)
	assert.True(t, called, "subscription must have been called")
}

func TestRegistry_delete(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(0xb0baFEDD))
	d := &mockDialer{dial: make(chan Conn)}
	r := NewRegistry(func(*Peer) {}, d)

	addr := wallet.NewRandomAddress(rng)
	p := r.Get(addr)
	p2, _ := r.find(addr)
	assert.Equal(t, p, p2)

	r.delete(p2)
	p2, _ = r.find(addr)
	assert.Nil(t, p2)

	assert.Panics(t, func() { r.delete(p) }, "double delete must panic")
}
