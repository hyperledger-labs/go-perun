// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package tcp implements a direct tcp implementation and fulfills
// the io.ReadWriteCloser interface.
package tcp

import (
	"net"
	"reflect"
	"testing"

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
		{"InvalidPort", args{host: "localhost", port: "70000"}, Connection{nil}, true},
		{"InvalidHost", args{host: "255.255.255.256", port: "1234"}, Connection{nil}, true},
		{"NoListenerRunning", args{host: "localhost", port: "1234"}, Connection{nil}, true},
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

func TestNewTCPListener(t *testing.T) {
	t.Parallel()
	type args struct {
		host string
		port string
	}
	tests := []struct {
		name    string
		args    args
		want    *Listener
		wantErr bool
	}{
		{"InvalidPort", args{host: "localhost", port: "70000"}, &Listener{}, true},
		{"InvalidHost", args{host: "255.255.255.256", port: "1234"}, &Listener{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewListener(tt.args.host, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTCPListener() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTCPListener() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListener(t *testing.T) {
	l := newTestListener(t)
	connClient, err := Connect(host, port)
	assert.Nil(t, err, "Connecting to a running Listener should not fail")
	connListener := <-l.Incoming
	// Client sends data to Listener
	data := "DATADATA"
	n, err := connClient.Write([]byte(data))
	assert.Nil(t, err, "Write to valid connection should not fail")
	assert.Equal(t, len(data), n, "Should have written len(data) bytes")
	buffer := make([]byte, 1024)
	n, err = connListener.Read(buffer)
	assert.Nil(t, err, "Reading from established channel should not fail")
	assert.Equal(t, len(data), n, "Should receive as much bytes as previously send")
	assert.Equal(t, []byte(data), buffer[:n], "Receiving should produce same data as previously send")
	// Listener sends data to client
	data = "DATADATADATADATA"
	n, err = connListener.Write([]byte(data))
	assert.Nil(t, err, "Write to valid connection should not fail")
	assert.Equal(t, len(data), n, "Should have written len(data) bytes")
	buffer = make([]byte, 1024)
	n, err = connClient.Read(buffer)
	assert.Nil(t, err, "Reading from established channel should not fail")
	assert.Equal(t, len(data), n, "Should receive as much bytes as previously send")
	assert.Equal(t, []byte(data), buffer[:n], "Receiving should produce same data as previously send")
	_, err = net.Listen("udp", host+":"+port)
	assert.NotNil(t, err, "Connecting with wrong protocol should fail")
	// Closing the connections
	err = l.Close()
	assert.Nil(t, err, "Closing of a Listener should not fail")
	err = connClient.Close()
	assert.Nil(t, err, "Closing of a client should not fail")
}

func TestDoubleConnect(t *testing.T) {
	_server, err := NewListener(host, port)
	defer _server.Close()
	assert.Nil(t, err, "Creating a TCPListener should not fail")
	listener, err := NewListener(host, port)
	assert.NotNil(t, err, "Creating a TCPListener on already used port should fail")
	err = listener.Close()
	assert.NotNil(t, err, "Closing of invalid Listener should fail")
}

func newTestListener(t *testing.T) *Listener {
	l, err := NewListener(host, port)
	assert.Nil(t, err, "Creating a TCPListener should not fail")
	return l
}
