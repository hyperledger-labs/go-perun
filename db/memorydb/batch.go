// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package memorydb

import (
	"github.com/pkg/errors"
)

type Batch struct {
	db      *Database
	writes  map[string]string
	deletes map[string]struct{}
}

func (this *Batch) Put(key string, value string) error {
	delete(this.deletes, key)
	this.writes[key] = value
	return nil
}

func (this *Batch) PutBytes(key string, value []byte) error {
	return this.Put(key, string(value))
}

func (this *Batch) Delete(key string) error {
	delete(this.writes, key)
	this.deletes[key] = struct{}{}
	return nil
}

func (this *Batch) Len() uint {
	return uint(len(this.writes) + len(this.deletes))
}

func (this *Batch) ValueSize() uint {
	bytes := 0

	for _, value := range this.writes {
		bytes += len(value)
	}

	return uint(bytes)
}

func (this *Batch) Apply() error {
	for key, value := range this.writes {
		err := this.db.Put(key, value)
		if err != nil {
			return errors.Wrap(err, "Failed to put entry.")
		}
	}

	for key := range this.deletes {
		err := this.db.Delete(key)
		if err != nil {
			return errors.Wrap(err, "Failed to delete entry.")
		}
	}
	return nil
}

func (this *Batch) Reset() {
	this.writes = make(map[string]string)
	this.deletes = make(map[string]struct{})
}
