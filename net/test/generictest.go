// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/net/test"

import (
	"testing"
	"net"

	"github.com/stretchr/testify/assert"
)

// ListenerFactory should create a new listener.
type ListenerFactory func() (net.Listener, error)

// Dialer should connect to a peer.
type Dialer func() (net.Conn, error)

// Setup provides two methods to the generic tests.
// The methods should create a new listener, and connect to a peer.
type Setup struct {
	ListenerFactory ListenerFactory
	Dialer   Dialer
}

// GenericListenerTest tests generic functionality of connecting and disconnecting of client and server.
func GenericListenerTest(t *testing.T, s *Setup) {
	// Starting a new listener
	l, err := s.ListenerFactory()
	assert.Nil(t, err, "Starting a listener should not fail")
	// Accepting a new connection
	accept := make(chan net.Conn)
	go func() {
		conn, err := l.Accept()
		assert.Nil(t, err, "Accepting a connection should not fail")
		accept <- conn
	}()
	// Connecting to the listener
	connClient, err := s.Dialer()
	connListener := <-accept
	// Client sends data to Listener
	data := "DATADATA"
	n, err := connClient.Write([]byte(data))
	assert.Nil(t, err, "Write to valid connection should not fail")
	assert.Equal(t, len(data), n, "Should have written len(data) bytes")
	buffer := make([]byte, 1024)
	n, err = connListener.Read(buffer)
	assert.Nil(t, err, "Reading from established channel should not fail")
	assert.Equal(t, len(data), n, "Should receive as many bytes as previously sent")
	assert.Equal(t, []byte(data), buffer[:n], "Receiving should produce same data as previously sent")
	// Listener sends data to client
	data = "DATADATADATADATA"
	n, err = connListener.Write([]byte(data))
	assert.Nil(t, err, "Write to valid connection should not fail")
	assert.Equal(t, len(data), n, "Should have written len(data) bytes")
	buffer = make([]byte, 1024)
	n, err = connClient.Read(buffer)
	assert.Nil(t, err, "Reading from established channel should not fail")
	assert.Equal(t, len(data), n, "Should receive as many bytes as previously sent")
	assert.Equal(t, []byte(data), buffer[:n], "Receiving should produce same data as previously sent")
	// Closing the connections
	err = l.Close()
	assert.Nil(t, err, "Closing of a listener should not fail")
	err = connClient.Close()
	assert.Nil(t, err, "Closing of a client should not fail")
	// Double closing
	err = l.Close()
	assert.NotNil(t, err, "Closing of an already closed listener should fail")
	err = connClient.Close()
	assert.NotNil(t, err, "Closing of an already closed client should fail")
}

// GenericDoubleConnectTest tests that creating a listener twice should fail.
func GenericDoubleConnectTest(t *testing.T, s *Setup) {
	_server, err := s.ListenerFactory()
	defer _server.Close()
	assert.Nil(t, err, "Creating a listener should not fail")
	_, err = s.ListenerFactory()
	assert.NotNil(t, err, "Creating a listener on already used address should fail")
}
