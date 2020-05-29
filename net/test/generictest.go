// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

// Package test provides generic tests for network connections.
// It tests basic functionality that every net.Conn and net.Listener should implement.
// For the definition of net.Conn and net.Listener see https://golang.org/pkg/net/.
//
// Usage:
// 		Provide a Setup struct to the GenericXTest functions.
// 		The setup struct must contain a function to create a listener/server
// 		and a function to connect to this listener/server.
package test // import "perun.network/go-perun/net/test"

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	requirepkg "github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/test"
)

// ListenerFactory should create a new listener.
type ListenerFactory func() (net.Listener, error)

// Dialer should connect to a peer.
type Dialer func() (net.Conn, error)

// Setup provides two methods to the generic tests.
// The methods should create a new listener, and connect to a peer.
type Setup struct {
	ListenerFactory ListenerFactory
	Dialer          Dialer
}

// GenericListenerTest tests generic functionality of connecting and disconnecting of client and server.
func GenericListenerTest(t *testing.T, s *Setup) {
	assert := assert.New(t)
	require := requirepkg.New(t)

	// Starting a new listener
	l, err := s.ListenerFactory()
	require.NoError(err, "Starting a listener should not fail")

	// Accepting a new connection
	accept := make(chan net.Conn)
	ct := test.NewConcurrent(t)
	go ct.Stage("accept", func(rt requirepkg.TestingT) {
		conn, aErr := l.Accept()
		requirepkg.NoError(rt, aErr, "Accepting a connection should not fail")
		accept <- conn
	})

	// Connecting to the listener
	connClient, err := s.Dialer()
	ct.Stage("errcheck", func(rt requirepkg.TestingT) {
		requirepkg.NoError(rt, err)
	})
	connListener := <-accept
	ct.Wait("accept")
	require.NotNil(connListener)

	// Client sends data to Listener
	data := "DATADATA"
	n, err := connClient.Write([]byte(data))
	require.NoError(err, "Write to valid connection should not fail")
	assert.Equal(len(data), n, "Should have written len(data) bytes")
	buffer := make([]byte, 1024)
	n, err = connListener.Read(buffer)
	require.NoError(err, "Reading from established channel should not fail")
	assert.Equal(len(data), n, "Should receive as many bytes as previously sent")
	assert.Equal([]byte(data), buffer[:n], "Receiving should produce same data as previously sent")

	// Listener sends data to client
	data = "DATADATADATADATA"
	n, err = connListener.Write([]byte(data))
	require.NoError(err, "Write to valid connection should not fail")
	assert.Equal(len(data), n, "Should have written len(data) bytes")
	buffer = make([]byte, 1024)
	n, err = connClient.Read(buffer)
	require.NoError(err, "Reading from established channel should not fail")
	assert.Equal(len(data), n, "Should receive as many bytes as previously sent")
	assert.Equal([]byte(data), buffer[:n], "Receiving should produce same data as previously sent")

	// Closing the connections
	err = l.Close()
	assert.NoError(err, "Closing of a listener should not fail")
	err = connClient.Close()
	assert.NoError(err, "Closing of a client should not fail")

	// Double closing
	err = l.Close()
	assert.Error(err, "Closing of an already closed listener should fail")
	err = connClient.Close()
	assert.Error(err, "Closing of an already closed client should fail")
}

// GenericDoubleConnectTest tests that creating a listener twice should fail.
func GenericDoubleConnectTest(t *testing.T, s *Setup) {
	_server, err := s.ListenerFactory()
	defer _server.Close()
	assert.NoError(t, err, "Creating a listener should not fail")
	_, err = s.ListenerFactory()
	assert.Error(t, err, "Creating a listener on already used address should fail")
}
