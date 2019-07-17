// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package tcp implements a direct tcp implementation and fulfills
// the io.ReadWriteCloser interface.
package tcp

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	host = "localhost"
	port = "13904"
)

func TestConnect(t *testing.T) {
	t.Parallel()
	type args struct {
		host string
		port string
	}
	tests := []struct {
		name    string
		args    args
		want    Connection
		wantErr bool
	}{
		{"InvalidPort", args{host: "localhost", port: "70000"}, Connection{nil, nil}, true},
		{"InvalidHost", args{host: "255.255.255.256", port: "1234"}, Connection{nil, nil}, true},
		{"NoServerRunning", args{host: "localhost", port: "1234"}, Connection{nil, nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Connect(tt.args.host, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Connect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTCPServer(t *testing.T) {
	t.Parallel()
	type args struct {
		host string
		port string
	}
	tests := []struct {
		name    string
		args    args
		want    *Server
		wantErr bool
	}{
		{"InvalidPort", args{host: "localhost", port: "70000"}, &Server{}, true},
		{"InvalidHost", args{host: "255.255.255.256", port: "1234"}, &Server{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewServer(tt.args.host, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTCPServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTCPServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConnectionError(t *testing.T) {
	t.Parallel()
	conErr := ConnectionError(nil)
	for i := 0; i < 10; i++ {
		err := errors.New(strconv.Itoa(i))
		conErr = append(conErr, err)
	}
	errorString := "0,1,2,3,4,5,6,7,8,9"
	assert.Equal(t, conErr.Error(), errorString, "Error messages should be equal")
}

func TestServer(t *testing.T) {
	server := newTestServer(t)
	connClient, err := Connect(host, port)
	assert.Nil(t, err, "Connecting to a running server should not fail")
	connServerChan := <-server.ConnChan
	connections := server.Connections()
	assert.Equal(t, 1, len(connections), "Server should have one connection")
	connServer := connections[0]
	assert.Equal(t, connServer, connServerChan, "Connections should be equal")
	// Client sends data to server
	data := "DATADATA"
	n, err := connClient.Write([]byte(data))
	assert.Nil(t, err, "Write to valid connection should not fail")
	assert.Equal(t, len(data), n, "Should have written len(data) bytes")
	buffer := make([]byte, 1024)
	n, err = connServer.Read(buffer)
	assert.Nil(t, err, "Reading from established channel should not fail")
	assert.Equal(t, len(data), n, "Should receive as much bytes as previously send")
	assert.Equal(t, []byte(data), buffer[:n], "Receiving should produce same data as previously send")
	// Server sends data to client
	data = "DATADATADATADATA"
	n, err = connServer.Write([]byte(data))
	assert.Nil(t, err, "Write to valid connection should not fail")
	assert.Equal(t, len(data), n, "Should have written len(data) bytes")
	buffer = make([]byte, 1024)
	n, err = connClient.Read(buffer)
	assert.Nil(t, err, "Reading from established channel should not fail")
	assert.Equal(t, len(data), n, "Should receive as much bytes as previously send")
	assert.Equal(t, []byte(data), buffer[:n], "Receiving should produce same data as previously send")
	// Closing the connections
	err = server.Close()
	assert.Nil(t, err, "Closing of a server should not fail")
	err = connClient.Close()
	assert.Nil(t, err, "Closing of a client should not fail")
	assert.Equal(t, 0, len(server.Connections()), "Server should have zero connections")
}

func TestDoubleConnect(t *testing.T) {
	_, err := NewServer(host, port)
	assert.Nil(t, err, "Creating a TCPServer should not fail")
	server, err := NewServer(host, port)
	assert.NotNil(t, err, "Creating a TCPServer on already used port should fail")
	err = server.Close()
	assert.NotNil(t, err, "Closing of invalid server should fail")
}

func newTestServer(t *testing.T) *Server {
	server, err := NewServer(host, port)
	assert.Nil(t, err, "Creating a TCPServer should not fail")
	return server
}
