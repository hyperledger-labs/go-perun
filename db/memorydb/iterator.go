// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package memorydb

import (
	"log"
)

type Iterator struct {
	next   int
	keys   []string
	values [][]byte
}

func (this *Iterator) Next() bool {
	if this.next <= len(this.keys) {
		this.next++
		return true
	} else {
		return false
	}
}

func (this *Iterator) Error() error {
	if this.next == 0 {
		log.Fatalln("Iterator.Error() accessed before Next().")
	}
	return nil
}

func (this *Iterator) Key() []byte {
	if this.next == 0 {
		log.Fatalln("Iterator.Key() accessed before Next().")
	}

	if this.next > len(this.keys) {
		return nil
	} else {
		return []byte(this.values[this.next-1])
	}
}

func (this *Iterator) Value() []byte {
	if this.next == 0 {
		log.Fatalln("Iterator.Value() accessed before Next().")
	}

	if this.next > len(this.keys) {
		return nil
	} else {
		return this.values[this.next-1]
	}
}

func (this *Iterator) Release() {
	this.next = 0
	this.keys = nil
	this.values = nil
}
