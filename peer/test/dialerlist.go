// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"sync"

	"github.com/pkg/errors"
)

type dialerList struct {
	mutex   sync.Mutex
	entries []*Dialer
}

func (l *dialerList) insert(dialer *Dialer) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.entries = append(l.entries, dialer)
}

func (l *dialerList) erase(dialer *Dialer) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for i, d := range l.entries {
		if d == dialer {
			l.entries[i] = l.entries[len(l.entries)-1]
			l.entries = l.entries[:len(l.entries)-1]
			return nil
		}
	}

	return errors.New("dialer does not exist")
}

func (l *dialerList) clear() []*Dialer {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	ret := l.entries
	l.entries = nil
	return ret
}
