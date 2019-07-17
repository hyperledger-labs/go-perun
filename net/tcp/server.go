// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package tcp implements a direct tcp implementation and fulfills
// the io.ReadWriteCloser interface.
package tcp // import "perun.network/go-perun/net/tcp"

import (
	"net"

	"github.com/pkg/errors"
	_log "perun.network/go-perun/log"
)

var log = _log.Log

const (
	queueSize = 10
)

// Listener represents a listener to a peer.
// A Listener is created with NewListener and binds to a port.
// If someone connects to the port, the Listener creates a new connection and stores it in incoming.
type Listener struct {
	listener net.Listener
	// Users can use the connection channel to get notified of new incoming connections.
	Incoming chan *Connection
	close    chan struct{}
}

// NewListener initializes a new tcp Listener and listens to incomming connections.
func NewListener(host, port string) (*Listener, error) {
	log.Info("Creating a new TCP Listener listening on " + host + ":" + port)
	socket, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		log.Warn("Could not create TCP Listener")
		return &Listener{}, errors.Wrap(err, "error trying to open connection on "+host+":"+port)
	}
	listener := &Listener{
		listener: socket,
		Incoming: make(chan *Connection, queueSize),
		close:    make(chan struct{}),
	}

	go listener.acceptIncomingConnections()
	return listener, nil
}

// Close closes all connections of the Listener.
func (s *Listener) Close() error {
	log.Debug("Closing all connections of Listener")
	if s.listener == nil {
		return errors.New("Listener has no valid listener")
	}
	close(s.close)
	return s.listener.Close()
}

func (s *Listener) acceptIncomingConnections() {
	for {
		c, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.close:
				log.Debugf("Closing Listener")
				close(s.Incoming)
				return
			default:
				log.Warnf("Incoming connection failed with error: ", err)
				continue
			}
		}
		conn := Connection{c}
		s.Incoming <- &conn
		log.Debugf("Accepted connection from peer ", conn.RemoteAddr().String())
	}
}
