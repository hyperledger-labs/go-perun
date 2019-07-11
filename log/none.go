// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package log

// None is a logger that doesn't do anything
var None Logger = &none{}

type none struct{}

func (none) Printf(string, ...interface{}) {}
func (none) Print(...interface{})          {}
func (none) Println(...interface{})        {}
func (none) Fatalf(string, ...interface{}) {}
func (none) Fatal(...interface{})          {}
func (none) Fatalln(...interface{})        {}
func (none) Panicf(string, ...interface{}) {}
func (none) Panic(...interface{})          {}
func (none) Panicln(...interface{})        {}
func (none) Tracef(string, ...interface{}) {}
func (none) Debugf(string, ...interface{}) {}
func (none) Infof(string, ...interface{})  {}
func (none) Warnf(string, ...interface{})  {}
func (none) Errorf(string, ...interface{}) {}
func (none) Trace(...interface{})          {}
func (none) Debug(...interface{})          {}
func (none) Info(...interface{})           {}
func (none) Warn(...interface{})           {}
func (none) Error(...interface{})          {}
func (none) Traceln(...interface{})        {}
func (none) Debugln(...interface{})        {}
func (none) Infoln(...interface{})         {}
func (none) Warnln(...interface{})         {}
func (none) Errorln(...interface{})        {}

func (n *none) WithField(key string, value interface{}) Logger {
	return n
}

func (n *none) WithFields(Fields) Logger {
	return n
}

func (n *none) WithError(error) Logger {
	return n
}
