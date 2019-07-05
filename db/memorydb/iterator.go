// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package memorydb

type Iterator struct {
	next   int
	keys   []string
	values []string
}

func (this *Iterator) Next() bool {
	if this.next < len(this.keys) {
		this.next++
		return true
	} else {
		return false
	}
}

func (this *Iterator) Key() string {
	if this.next == 0 {
		panic("Iterator.Key() accessed before Next() or after Close().")
	}

	if this.next > len(this.keys) {
		return ""
	} else {
		return this.values[this.next-1]
	}
}

func (this *Iterator) Value() string {
	if this.next == 0 {
		panic("Iterator.Value() accessed before Next() or after Close().")
	}

	if this.next > len(this.keys) {
		return ""
	} else {
		return this.values[this.next-1]
	}
}

func (this *Iterator) ValueBytes() []byte {
	return []byte(this.Value())
}

func (this *Iterator) Close() error {
	this.next = 0
	this.keys = nil
	this.values = nil
	return nil
}
